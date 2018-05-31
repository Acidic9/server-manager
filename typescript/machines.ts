/// <reference path="./machine.ts" />

var selectedMachine: number;

$(document).ready(function() {
	$(document).on('change', "input:checkbox.machine-checkbox", onMachineCheckboxChange);

	$(".select-all-checkbox").click(onSelectAllCheckboxClick);
	$("#remove-machine-btn").click(onRemoveMachineClick);
});

function onMachineCheckboxChange(event: JQueryEventObject): void {
	var cbox = $(event.target);
	var isChecked = cbox.is(":checked");
	$("input:checkbox.machine-checkbox").prop("checked", false);
	$("tr.machine-row").removeClass("is-selected");
	cbox.prop("checked", isChecked);
	var tr = cbox.closest("tr.machine-row");
	var machineID = tr.data("id");
	if (isChecked) {
		selectedMachine = machineID;
		tr.addClass("is-selected");
	}

	var machineBtns = $(".machine-buttons a.button").addClass("is-outlined");
	$(".select-all-checkbox").prop("checked", false);
	if ($("input:checkbox:checked.machine-checkbox").length > 0) {
		$(".select-all-checkbox").prop("checked", true);
		machineBtns.removeClass("is-outlined");
	}
}

function onSelectAllCheckboxClick(event: JQueryEventObject): void {
	if ($(event.target).is(":checked")) {
		$("input:checkbox.machine-checkbox").prop("checked", true);
	} else {
		$("input:checkbox.machine-checkbox").prop("checked", false);
	}
	$("input:checkbox.machine-checkbox").trigger("change");
}

function onRemoveMachineClick(event: JQueryEventObject): void {
	if (selectedMachine) {
		var machine = new Machine(selectedMachine);
		machine
			.delete(false)
			.done(resp => {
				if (resp.error)
					console.warn(resp.error);
				
				if (!resp.success) {
					showError(resp.error);
					return
				}
				$("#machine-"+machine.id).remove();
			});
	}
}