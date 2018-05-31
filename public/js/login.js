var isMobile = false;
$(document).ready(function () {
    updateIsMobile();
    $(window).resize(updateIsMobile);
    $("#login-form").submit(submitLoginForm);
});
function updateIsMobile() {
    if ($('.screen-size').width() == 768)
        isMobile = true;
    else
        isMobile = false;
}
function submitLoginForm(e) {
    e.preventDefault();
    if (!validateForm()) {
        if (!isMobile)
            $(".login-container").effect("shake", { distance: 2, times: 3 });
        return;
    }
    hideValidation();
    disableForm();
    var postData = {
        username: $("#username").val().replace(/^\s+|\s+$/g, ""),
        password: $("#password").val(),
    };
    $.post("/user/login", postData, function (resp) {
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
                alertify.alert(errorMsgs[resp.error], function () {
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
    }, "json").fail(function () {
        alertify.error(errorMsgs["default"]);
    }).always(function () {
        enableForm();
    });
}
function validateForm() {
    hideValidation();
    var username = $("#username").val().split(' ').join(''), password = $("#password").val();
    if (username.length == 0) {
        $(".username-field .help").html(errorMsgs["empty-username"]).show();
        $("#username").focus();
        return false;
    }
    return true;
}
function hideValidation() {
    $(".username-field .help").html("").hide();
    $(".password-field .help").html("").hide();
}
function disableForm() {
    $("#username").prop("disabled", true);
    $("#password").prop("disabled", true);
    $("#submit-btn").prop("disabled", true);
}
function enableForm() {
    $("#username").prop("disabled", false);
    $("#password").prop("disabled", false);
    $("#submit-btn").prop("disabled", false);
}
