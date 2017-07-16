var selectedMachines = [];

$(document).ready(function() {
	$(document).on('change', "input:checkbox.machine-checkbox", onMachineCheckboxChange);

	$(".select-all-checkbox").click(onSelectAllCheckboxClick);
});

function onMachineCheckboxChange() {
	var cbox = $(this);
	var tr = cbox.closest("tr");
	if (cbox.is(":checked")) {
		selectedMachines.pushIfNotExist(tr.attr("id"));
		cbox.closest("tr").addClass("is-selected");
	} else {
		selectedMachines.delete(tr.attr("id"));
		cbox.prop("checked", false);
		cbox.closest("tr").removeClass("is-selected");
	}

	var machineBtns = $(".machine-buttons a.button").addClass("is-outlined");
	$(".select-all-checkbox").prop("checked", false);
	if ($("input:checkbox:checked.machine-checkbox").length > 0) {
		$(".select-all-checkbox").prop("checked", true);
		machineBtns.removeClass("is-outlined");
	}
}

function onSelectAllCheckboxClick() {
	if ($(this).is(":checked")) {
		$("input:checkbox.machine-checkbox").prop("checked", true);
	} else {
		$("input:checkbox.machine-checkbox").prop("checked", false);
	}
	$("input:checkbox.machine-checkbox").trigger("change");
}