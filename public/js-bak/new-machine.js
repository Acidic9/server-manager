$(document).ready(function() {
	$("#new-machine-form").submit(submitNewMachineForm);
	$("#test-connection").click(testMachineConnection);
});

function submitNewMachineForm(e) {
	e.preventDefault();

	$("#submit-new-machine-form").addClass("is-loading");
	$("#new-machine-form").attr("disabled", true);

	if (!validateForm()) {
		$("#submit-new-machine-form").removeClass("is-loading");
		$("#new-machine-form").attr("disabled", false);
		return;
	}

	var title = $("#title").val();
	var address = $("#address").val().split(' ').join('');
	var port = $("#port").val().split(' ').join('');
	var username = $("#username").val().split(' ').join('');
	var password = $("#password").val();

	var data = {
		"title": title,
		"addr": address,
		"port": port,
		"username": username,
		"password": password
	};

	$.post("/machine/add", data, function(resp) {
		if (resp.error)
			console.warn(resp.error);

		if (!resp.success) {
			showError(resp.error);
			return;
		}

		window.location = "/machines";
	}, "json").always(function() {
		$("#submit-new-machine-form").removeClass("is-loading");
		$("#new-machine-form").attr("disabled", false);
	});
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
			$(".address-field .help").html(errorMsgs["invalid-ip-address"]).show();
			$("#address").focus();
			return false;

		case !/^\d{1,4}$/g.test(port):
			$(".port-field .help").html(errorMsgs["invalid-port"]).show();
			$("#port").focus();
			return false;

		case username.length == 0:
			$(".username-field .help").html(errorMsgs["empty-username"]).show();
			$("#username").focus();
			return false;
	}

	return true;
}

function testMachineConnection() {
	$("#test-connection").addClass("is-loading");
	$("#new-machine-form").attr("disabled", true);

	if (!validateForm()) {
		$("#test-connection").removeClass("is-loading");
		$("#new-machine-form").attr("disabled", false);
		return;
	}

	var title = $("#title").val();
	var address = $("#address").val().split(' ').join('');
	var port = $("#port").val().split(' ').join('');
	var username = $("#username").val().split(' ').join('');
	var password = $("#password").val();

	var data = {
		"addr": address,
		"port": port,
		"username": username,
		"password": password
	};

	$.get("/machine/test-connection", data, function(resp) {
		if (resp.error)
			console.warn(resp.error);

		if (!resp.success) {
			showError(resp.error);
			return;
		}

		alertify.success("Successfully established connection to remote machine");
	}, "json").always(function() {
		$("#test-connection").removeClass("is-loading");
		$("#new-machine-form").attr("disabled", false);
	});
}