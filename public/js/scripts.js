var errorMsgs = {
	default: "Something went wrong",
	emptyUsername: "You must enter a username",
	invalidIPAddress: "That IP address is not valid",
	invalidPort: "The port must be 1-4 digits",
	incorrectCredentials: "Incorrect login - Your username and password don't match",
	databaseError: "A database error occured",
	notLoggedIn: "You must be logged in",
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