class Machine {
	/**
	 * @param {number} id
	 * @param {string} title
	 * @param {string} addr
	 * @param {number|null} port
	 * @param {string} username
	 * @param {string|null} password
	 */
	constructor(id, title, addr, port, username, password) {
		this.id = id;
		this.title = title;
		this.addr = addr;
		this.port = port;
		this.username = username;
		this.password = password;
	}

	/**
	 * @param {boolean|null} async
	 * @return {jqXHR}
	 */
	delete(async = true) {
		var machine = this;
		var data = {"machineID": this.id};
		return $.ajax({
			type: "post",
			url: "/machine/delete",
			machine: machine,
			data: data,
			dataType: "json",
			async: async,
		});
	}
}