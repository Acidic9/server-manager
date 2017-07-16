$(document).ready(function() {
	$("#new-machine-form").submit(submitNewMachineForm);	
});

function submitNewMachineForm(e) {
	e.preventDefault();

	if (!validateForm()) {
		return;
	}

	var title = $("#title").val();
	var address = $("#address").val().split(' ').join('');
	var port = $("#port").val().split(' ').join('');
	var username = $("#username").val().split(' ').join('');
	var password = $("#password").val();

	var data = {
		"title": title,
		"address": address,
		"port": port,
		"username": username,
		"password": password
	};

	$.post("/machine/add", data,
		function(resp) {
			if (!resp.success) {
				alertify.error(errorMsgs.default);
				if (resp.error.length > 0)
					console.log(resp.error);

				return;
			}
			alert("Success")
		},
		"json"
	);
}

function validateForm() {
	var title = $("#title").val();
	var address = $("#address").val().split(' ').join('');
	var port = $("#port").val().split(' ').join('');
	var username = $("#username").val().split(' ').join('');
	var password = $("#password").val();

	$(".address-field .help").hide();
	$(".port-field .help").hide();
	$(".username-field .help").hide();

	switch (true) {
		case !/^\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b$/g.test(address):
			$(".address-field .help").html(errorMsgs.invalidIPAddress).show();
			$("#address").focus();
			return false;

		case !/^\d{1,4}$/g.test(port):
			$(".port-field .help").html(errorMsgs.invalidPort).show();
			$("#port").focus();
			return false;

		case username.length == 0:
			$(".username-field .help").html(errorMsgs.emptyUsername).show();
			$("#username").focus();
			return false;
	}

	return true;
}