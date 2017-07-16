#!/bin/bash

gameservername=$1
source "lgsm/config-lgsm/${gameservername}/_default.cfg"

fn_deps_detector(){
	# Checks if dependency is missing
	if [ "${tmuxcheck}" == "1" ]; then
		# Added for users compiling tmux from source to bypass check.
		depstatus=0
		deptocheck="tmux"
		unset tmuxcheck
	elif [ "${javacheck}" == "1" ]; then
		# Added for users using Oracle JRE to bypass check.
		depstatus=0
		deptocheck="${javaversion}"
		unset javacheck
	elif [ -n "$(command -v apt-get 2>/dev/null)" ]; then
		dpkg-query -W -f='${Status}' ${deptocheck} 2>/dev/null | grep -q -P '^install ok installed$'
		depstatus=$?
	elif [ -n "$(command -v yum 2>/dev/null)" ]; then
		yum -q list installed ${deptocheck} > /dev/null 2>&1
		depstatus=$?
	fi

	if [ "${depstatus}" == "0" ]; then
		# if dependency is found
		missingdep=0
	else
		# if dependency is not found
		missingdep=1
	fi

	# Missing dependencies are added to array_deps_missing
	if [ "${missingdep}" == "1" ]; then
		array_deps_missing+=("${deptocheck}")
	fi
}

fn_found_missing_deps(){
	if [ -n "$(command -v dpkg-query 2>/dev/null)" ]; then
		echo "manaul_install_deps_command=\"sudo dpkg --add-architecture i386; sudo apt-get update; sudo apt-get install ${array_deps_missing[@]}\""
	elif [ -n "$(command -v yum 2>/dev/null)" ]; then
		echo "manaul_install_deps_command=\"sudo yum install ${array_deps_missing[@]}\""
	fi
	if [ "${#array_deps_missing[@]}" != "0" ]; then
		sudo -v > /dev/null 2>&1
		if [ $? -eq 0 ]; then
			# Automatically install deps...
			if [ -n "$(command -v dpkg-query 2>/dev/null)" ]; then
				cmd="sudo dpkg --add-architecture i386; sudo apt-get update; sudo apt-get -y install ${array_deps_missing[@]}"
				eval ${cmd}
			elif [ -n "$(command -v yum 2>/dev/null)" ]; then
				cmd="sudo yum -y install ${array_deps_missing[@]}"
				eval ${cmd}
			fi
			if [ $? != 0 ]; then
				echo -e "missing_deps=\"${array_deps_missing[@]}\""
			else
				echo -e "missing_deps=\"\""
			fi
		else
			echo -e "missing_deps=\"${array_deps_missing[@]}\""
		fi
	else
		echo -e "missing_deps=\"${array_deps_missing[@]}\""
	fi
}

fn_check_loop(){
	# Loop though required depenencies
	for deptocheck in "${array_deps_required[@]}"
	do
		fn_deps_detector
	done

	# user to be informed of any missing dependencies
	fn_found_missing_deps
}

arch=$(uname -m)
kernel=$(uname -r)
if [ -n "$(command -v lsb_release)" ]; then
	distroname=$(lsb_release -s -d)
elif [ -f "/etc/os-release" ]; then
	distroname=$(grep PRETTY_NAME /etc/os-release | sed 's/PRETTY_NAME=//g' | tr -d '="')
elif [ -f "/etc/debian_version" ]; then
	distroname="Debian $(cat /etc/debian_version)"
elif [ -f "/etc/redhat-release" ]; then
	distroname=$(cat /etc/redhat-release)
else
	distroname="$(uname -s) $(uname -r)"
fi

if [ -f "/etc/os-release" ]; then
	distroversion=$(grep VERSION_ID /etc/os-release | tr -cd '[:digit:]')
elif [ -n "$(command -v yum)" ]; then
	distroversion=$(rpm -qa \*-release | grep -Ei "oracle|redhat|centos" | cut -d"-" -f3)
fi

## Glibc version
# e.g: 1.17
glibcversion="$(ldd --version | sed -n '1s/.* //p')"

## tmux version
# e.g: tmux 1.6
if [ -z "$(command -V tmux 2>/dev/null)" ]; then
	tmuxv="${red}NOT INSTALLED!${default}"
else
	if [ "$(tmux -V|sed "s/tmux //" | sed -n '1 p' | tr -cd '[:digit:]')" -lt "16" ] 2>/dev/null; then
		tmuxv="$(tmux -V) (>= 1.6 required for console log)"
	else
		tmuxv=$(tmux -V)
	fi
fi

# Check will only run if using apt-get or yum
if [ -n "$(command -v dpkg-query 2>/dev/null)" ]; then
	# Generate array of missing deps
	array_deps_missing=()

	# LinuxGSM requirements
	array_deps_required=( curl wget ca-certificates file bsdmainutils util-linux python bzip2 gzip unzip binutils )

	# All servers except ts3 require tmux
	if [ "${gamename}" != "TeamSpeak 3" ]; then
		if [ "$(command -v tmux 2>/dev/null)" ]||[ "$(which tmux 2>/dev/null)" ]||[ -f "/usr/bin/tmux" ]||[ -f "/bin/tmux" ]; then
			tmuxcheck=1 # Added for users compiling tmux from source to bypass check.
		else
			array_deps_required+=( tmux )
		fi
	fi

	# All servers except ts3,mumble,multitheftauto and minecraft servers require libstdc++6 and lib32gcc1
	if [ "${gamename}" != "TeamSpeak 3" ]&&[ "${gamename}" != "Mumble" ]&&[ "${engine}" != "lwjgl2" ]&&[ "${engine}" != "renderware" ]; then
		if [ "${arch}" == "x86_64" ]; then
			array_deps_required+=( lib32gcc1 libstdc++6:i386 )
		else
			array_deps_required+=( libstdc++6:i386 )
		fi
	fi

	# Game Specific requirements

	# Spark
	if [ "${engine}" ==  "spark" ]; then
		array_deps_required+=( speex:i386 libtbb2 )
	# 7 Days to Die
	elif [ "${gamename}" ==  "7 Days To Die" ]; then
		array_deps_required+=( telnet expect )
	# No More Room in Hell, Counter-Strike: Source and Garry's Mod
	elif [ "${gamename}" == "No More Room in Hell" ]||[ "${gamename}" == "Counter-Strike: Source" ]||[ "${gamename}" == "Garry's Mod" ]; then
		if [ "${arch}" == "x86_64" ]; then
			array_deps_required+=( lib32tinfo5 )
		else
			array_deps_required+=( libtinfo5 )
		fi
	# Brainbread 2 ,Don't Starve Together & Team Fortress 2
	elif [ "${gamename}" == "Brainbread 2" ]||[ "${gamename}" == "Don't Starve Together" ]||[ "${gamename}" == "Team Fortress 2" ]; then
		array_deps_required+=( libcurl4-gnutls-dev:i386 )
	# Battlefield: 1942
	elif [ "${gamename}" == "Battlefield: 1942" ]; then
		array_deps_required+=( libncurses5:i386 )
	# Call of Duty
	elif [ "${gamename}" == "Call of Duty" ]||[ "${gamename}" == "Call of Duty 2" ]; then
		array_deps_required+=( libstdc++5:i386 )
	# Project Zomboid and Minecraft
	elif [ "${engine}" ==  "projectzomboid" ]||[ "${engine}" == "lwjgl2" ]; then
		javaversion=$(java -version 2>&1 | grep "version")
		if [ -n "${javaversion}" ]; then
			javacheck=1 # Added for users using Oracle JRE to bypass the check.
		else
			array_deps_required+=( default-jre )
		fi
	# GoldenEye: Source
	elif [ "${gamename}" ==  "GoldenEye: Source" ]; then
		array_deps_required+=( zlib1g:i386 libldap-2.4-2:i386 )
	# Serious Sam 3: BFE
	elif [ "${gamename}" ==  "Serious Sam 3: BFE" ]; then
		array_deps_required+=( libxrandr2:i386 libglu1-mesa:i386 libxtst6:i386 libusb-1.0-0-dev:i386 libxxf86vm1:i386 libopenal1:i386 libssl1.0.0:i386 libgtk2.0-0:i386 libdbus-glib-1-2:i386 libnm-glib-dev:i386 )
	# Unreal Engine
	elif [ "${executable}" ==  "./ucc-bin" ]; then
		#UT2K4
		if [ -f "${executabledir}/ut2004-bin" ]; then
			array_deps_required+=( libsdl1.2debian libstdc++5:i386 bzip2 )
		#UT99
		else
			array_deps_required+=( libsdl1.2debian bzip2 )
		fi
	# Unreal Tournament
	elif [ "${gamename}" == "Unreal Tournament" ]; then
		array_deps_required+=( unzip )
	fi
	fn_check_loop
elif [ -n "$(command -v yum 2>/dev/null)" ]; then
	# Generate array of missing deps
	array_deps_missing=()

	# LinuxGSM requirements
	if [ "${distroversion}" == "6" ]; then
		array_deps_required=( curl wget util-linux-ng python file gzip bzip2 unzip )
	else
		array_deps_required=( curl wget util-linux python file gzip bzip2 unzip )
	fi

	# All servers except ts3 require tmux
	if [ "${gamename}" != "TeamSpeak 3" ]; then
		if [ "$(command -v tmux 2>/dev/null)" ]||[ "$(which tmux 2>/dev/null)" ]||[ -f "/usr/bin/tmux" ]||[ -f "/bin/tmux" ]; then
			tmuxcheck=1 # Added for users compiling tmux from source to bypass check.
		else
			array_deps_required+=( tmux )
		fi
	fi

	# All servers except ts3,mumble,multitheftauto and minecraft servers require glibc.i686 and libstdc++.i686
	if [ "${gamename}" != "TeamSpeak 3" ]&&[ "${gamename}" != "Mumble" ]&&[ "${engine}" != "lwjgl2" ]&&[ "${engine}" != "renderware" ]; then
		array_deps_required+=( glibc.i686 libstdc++.i686 )
	fi

	# Game Specific requirements

	# Spark
	if [ "${engine}" ==  "spark" ]; then
		array_deps_required+=( speex.i686 tbb.i686 )
	# 7 Days to Die
	elif [ "${gamename}" ==  "7 Days To Die" ]; then
		array_deps_required+=( telnet expect )
	# No More Room in Hell, Counter-Strike: Source and Garry's Mod
	elif [ "${gamename}" == "No More Room in Hell" ]||[ "${gamename}" == "Counter-Strike: Source" ]||[ "${gamename}" == "Garry's Mod" ]; then
		array_deps_required+=( ncurses-libs.i686 )
	# Brainbread 2, Don't Starve Together & Team Fortress 2
	elif [ "${gamename}" == "Brainbread 2" ]||[ "${gamename}" == "Don't Starve Together" ]||[ "${gamename}" == "Team Fortress 2" ]; then
		array_deps_required+=( libcurl.i686 )
	# Battlefield: 1942
	elif [ "${gamename}" == "Battlefield: 1942" ]; then
		array_deps_required+=( ncurses-libs.i686 )
	# Call of Duty
	elif [ "${gamename}" == "Call of Duty" ]||[ "${gamename}" == "Call of Duty 2" ]; then
		array_deps_required+=( compat-libstdc++-33.i686 )
	# Project Zomboid and Minecraft
	elif [ "${engine}" ==  "projectzomboid" ]||[ "${engine}" == "lwjgl2" ]; then
		javaversion=$(java -version 2>&1 | grep "version")
		if [ -n "${javaversion}" ]; then
			javacheck=1 # Added for users using Oracle JRE to bypass the check.
		else
			array_deps_required+=( java-1.8.0-openjdk )
		fi
	# GoldenEye: Source
	elif [ "${gamename}" ==  "GoldenEye: Source" ]; then
		array_deps_required+=( zlib.i686 openldap.i686 )
	# Unreal Engine
	elif [ "${executable}" ==  "./ucc-bin" ]; then
		#UT2K4
		if [ -f "${executabledir}/ut2004-bin" ]; then
			array_deps_required+=( compat-libstdc++-33.i686 SDL.i686 bzip2 )
		#UT99
		else
			array_deps_required+=( SDL.i686 bzip2 )
		fi
	# Unreal Tournament
	elif [ "${gamename}" == "Unreal Tournament" ]; then
		array_deps_required+=( unzip )
	fi
	fn_check_loop
fi