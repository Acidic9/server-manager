var errorMsgs = {
	"default": "Something went wrong",
	"empty-username": "You must enter a username",
	"invalid-ip-address": "That IP address is not valid",
	"empty-ip-address": "You must enter an IP address",
	"invalid-port": "The port must be between 1 and 9999",
	"incorrect-credentials": "Incorrect login - Your username and password don't match",
	"database-error": "A database error occured",
	"not-logged-in": "You must be logged in",
	"machine-connection-failed": "Failed to connect to the remote machine",
	"machine-does-not-exist": "Machine does not exist",
	"root-not-allowed": "Do not use root as the username"
}

function showError(error, showDefault, modify) {
	var errorMsg = errorMsgs[error];

	if (errorMsg) {
		if (modify)
			errorMsg = modify(errorMsg);
		
		alertify.error(errorMsg);
		return true;
	}

	if (showDefault === false)
		return false;

	errorMsg = errorMsgs["default"];

	if (modify)
		errorMsg = modify(errorMsgs["default"]);

	alertify.error(errorMsg);
	return true;
}

Array.prototype.pushIfNotExist = function(e) {
	var exists = false;
	for (var i = 0; i < this.length; i++) { 
        if (this[i] == e) {
			exists = true;
			break;
		}
    }

    if (!exists) {
        this.push(e);
    }
};

Array.prototype.delete = function(e) {
	for (var i = this.length-1; i >= 0; i--) { 
        if (this[i] == e) {
			this.splice(i, 1);
		}
    }
}

$(document).ready(function() {
	if ($ && $.fn && $.fn.select2)
		$(".select2").select2();
});