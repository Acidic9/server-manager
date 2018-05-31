declare var alertify: any;
declare var select2: any;

var errorMsgs: { [key:string]:string; } = {
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
	"root-not-allowed": "Do not use root as the username",
}

/**
 * Display a popup error to the user.
 * 
 * @param {string} error 
 * @param {boolean} [showDefault] 
 * @param {(para:string)=>string} [modify] 
 * @returns {boolean} 
 */
function showError(error: string, showDefault?: boolean, modify?: (para:string)=>string): boolean {
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

$(document).ready(function() {
	$.fn.ajaxStart
	if ($ && $.fn && $.fn.select2)
		$(".select2").select2();
});