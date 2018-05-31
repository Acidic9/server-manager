var isMobile: boolean = false;

$(document).ready(function() {
	// Mobile/Desktop detection.
	updateIsMobile();
	$(window).resize(updateIsMobile);

	// On submit login form.
	$("#login-form").submit(submitLoginForm);
});

/**
 * Update isMobile variable
 * 
 */
function updateIsMobile(): void {
	if ($('.screen-size').width() == 768)
		isMobile = true;
	else
		isMobile = false;
}

/**
 * Submit login form and login if correct details.
 * 
 * @param {JQuery.Event} e 
 * @returns {void} 
 */
function submitLoginForm(e: JQuery.Event): void {
	e.preventDefault(); // Prevent default form submission

	// Validate form
	if (!validateForm()) {
		if (!isMobile)
			$(".login-container").effect("shake", { distance: 2, times: 3 });
		return;
	}

	hideValidation();
	disableForm();

	// Post data to /user/login
	var postData = {
		username: ($("#username").val() as string).replace(/^\s+|\s+$/g, ""),
		password: $("#password").val(),
	}
	$.post("/user/login", postData,
		function(resp) {
			if (resp.error)
				console.warn(resp.error);

			if (resp.success == true) {
				window.location.href = "/";
				return;
			}

			if (!isMobile)
				$(".login-container").effect("shake", { distance: 4, times: 3 });

			switch (resp.error) {
				case "already-logged-in":
					alertify.alert(errorMsgs[resp.error], function() {
						window.location.href = "/";
					});
					break;

				case "empty-username":
					$(".username-field .help").html(errorMsgs[resp.error]).show();
					$("#username").focus();
					break;

				default:
					showError(resp.error);
					$("#username").focus();
					break;
			}
		},
		"json"
	).fail(function() {
		alertify.error(errorMsgs["default"]);
	}).always(function() {
		enableForm();
	});
}

/**
 * Validate form fields.
 * 
 * @returns {boolean} 
 */
function validateForm(): boolean {
	hideValidation();

	var username = ($("#username").val() as string).split(' ').join(''),
		password = $("#password").val();

	if (username.length == 0) {
		$(".username-field .help").html(errorMsgs["empty-username"]).show();
		$("#username").focus();
		return false;
	}

	return true;
}

/**
 * Hide all validation errors.
 * 
 */
function hideValidation(): void {
	$(".username-field .help").html("").hide();
	$(".password-field .help").html("").hide();
}
/**
 * Disable all form inputs and buttons.
 * 
 */
function disableForm() {
	$("#username").prop("disabled", true);
	$("#password").prop("disabled", true);
	$("#submit-btn").prop("disabled", true);
}
/**
 * Enable all form inputs and buttons.
 * 
 */
function enableForm() {
	$("#username").prop("disabled", false);
	$("#password").prop("disabled", false);
	$("#submit-btn").prop("disabled", false);
}