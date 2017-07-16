var selectedServers = [];
var servers = [];
var serverStatusWs;

reloadMachines();

$(document).ready(function() {
	$(document).on('change', "input:checkbox.server-checkbox", onServerCheckboxChange);
	
	serverStatusWs = new WebSocket("ws://"+window.location.host+"/server/status")
	serverStatusWs.onmessage = onServerStatusMsg;

	$(".select-all-checkbox").click(onSelectAllCheckboxClick);

	$("#start-server-btn").data("can-click", true);
	$("#start-server-btn").click(startServer);

	$("#stop-server-btn").data("can-click", true);
	$("#stop-server-btn").click(stopServer);
});

function reloadMachines() {
	$(".servers-table").each(function() {
		var machineID = $(this).data("machine-id");
		var tag = $(this).prev(".tag");
		var serversTable = $(this);
		tag.hide();
		serversTable.hide();
		var offlineCount = 0;
		var totalServers = 0;
		var isShown = false;
		$.each(servers, function(i, s) { 
			if (s.MachineID == machineID) {
				totalServers++;
				if (s.Status == "offline")
					offlineCount++;
				if (!isShown) {
					isShown = true;
					tag.show();
					serversTable.show();
				}
			}
		});
		tag.addClass("is-primary");
		tag.removeClass("is-danger");
		if (offlineCount > 0 && offlineCount == totalServers) {
			// Machine is offline
			tag.removeClass("is-primary");
			tag.addClass("is-danger");
		}
	});
}

function onServerCheckboxChange() {
	var cbox = $(this);
	var tr = cbox.closest("tr");
	if (cbox.is(":checked")) {
		selectedServers.pushIfNotExist(tr.attr("id"));
		cbox.closest("tr").addClass("is-selected");
	} else {
		selectedServers.delete(tr.attr("id"));
		cbox.prop("checked", false);
		cbox.closest("tr").removeClass("is-selected");
	}

	var serverBtns = $(".server-buttons a.button").addClass("is-outlined");
	$(".select-all-checkbox").prop("checked", false);
	if ($("input:checkbox:checked.server-checkbox").length > 0) {
		$(".select-all-checkbox").prop("checked", true);
		serverBtns.removeClass("is-outlined");
	}
}

function onServerStatusMsg(msg) {
	var server = JSON.parse(msg.data);
	var found = false;
	var updated = false;
	$.each(servers, function(sid, oldServer) { 
		if (server.ID == oldServer.ID && server.MachineID == oldServer.MachineID) {
			found = true;
			updated = true;
			servers[sid] = server;
		}
	});

	if (!found) {
		servers.push(server);
		updated = true;
	}

	if (updated) {
		var status = '<span class="tag is-primary server-status">'+server.Status+'</span>';
		switch (server.Status) {
			case "...":
				status = '<span class="tag server-status"><i class="fa fa-refresh spinning" style="color: #4285F4" aria-hidden="true"></i></span>';
				break;

			case "starting":
			case "stopping":
				status = '<span class="tag is-warning server-status"><i class="fa fa-refresh spinning mr4" aria-hidden="true"></i>'+server.Status+'</span>';
				break;

			case "online":
			case "running":
				status = '<span class="tag is-success server-status">'+server.Status+'</span>';
				break;

			case "offline":
			case "stopped":
				status = '<span class="tag is-danger server-status">'+server.Status+'</span>';
				break;

			case "installing":
				status = '<span class="tag is-info server-status"><i class="fa fa-spinner spinning mr4" aria-hidden="true"></i>'+server.Status+'</span>';
				break;

			case "":
				status = '<span class="tag is-warning server-status">idle</span>';
				break;
		}

		var tr = `<td class="small-column">
					<input type="checkbox" class="server-checkbox">
				</td>
				<td class="small-column server-status-col">`+status+`</td>
				<td>
					<small><i>No Name Set</i></small>
				</td>
				<td class="server-game-col">`+server.Game+`</td>
				<td>
					<a title="Manage"><i class="fa fa-wrench" aria-hidden="true" style="padding: 0 10px; margin-top: 2px"></i></a>
				</td>`;

		var found = false;
		$(".server-row").each(function() {
			var serverID = $(this).data("server-id");
			var machineID = $(this).data("machine-id");
			if (server.ID == serverID && server.MachineID == machineID) {
				found = true;
				var statusText = $(this).find(".server-status-col").text();
				var gameText = $(this).find(".server-game-col").text();
				if (statusText != server.Status) {
					$(this).find(".server-status-col").html(status);
				}
				if (gameText != server.Game) {
					$(this).find(".server-game-col").html(server.Game);
				}
				return false;
			}
		});

		if (!found) {
			$(".servers-table").each(function() {
				var machineID = $(this).data("machine-id");
				if (machineID == server.MachineID) {
					$(this).find("tbody").append(`<tr id="server-`+server.ID+`-`+server.MachineID+`" class="server-row" data-server-id="`+server.ID+`" data-machine-id="`+server.MachineID+`">`+tr+`</tr>`);
				}
			});
		}

		reloadMachines();
	};
}

function deselectServers() {
	for (var i = selectedServers.length-1; i >= 0; i--) {
		var s = $("#"+selectedServers[i]);
		selectedServers.splice(i, 1);
		s.removeClass("is-selected");
		s.find("input:checkbox").prop("checked", false);
	}
}

function setServerRowLoading(tr, loading) {
	tr.find("input[type=checkbox]").parent().find("i").remove();
	if (loading) {
		tr.find("input[type=checkbox]").hide();
		tr.find("input[type=checkbox]").parent().append(`<i class="fa fa-refresh spinning" style="margin-top: 4px; color: #4285F4" aria-hidden="true"></i>`);
		return;
	}
	tr.find("input[type=checkbox]").show();
	tr.find("input[type=checkbox]").parent().find("i").remove();
}

function onSelectAllCheckboxClick() {
	if ($(this).is(":checked")) {
		$("input:checkbox.server-checkbox").prop("checked", true);
	} else {
		$("input:checkbox.server-checkbox").prop("checked", false);
	}
	$("input:checkbox.server-checkbox").trigger("change");
}

function startServer() {
	if (selectedServers.length == 0 || !$("#start-server-btn").data("can-click"))
		return;

	$("#start-server-btn").data("can-click", false);
	setTimeout(function() {
		$("#start-server-btn").data("can-click", true);
	}, 500);

	for (var i = selectedServers.length-1; i >= 0; i--) {
		var s = $("#"+selectedServers[i]);
		setServerRowLoading(s, true);
		var serverID = s.data("server-id");
		var machineID = s.data("machine-id");
		$.ajax({
			type: "post",
			url: "/server/start",
			server: s,
			data: {"serverID": serverID, "machineID": machineID},
			dataType: "json",
			async: true,
			success: function(resp) {
				if (!resp.success) {
					var outputLink = "";
					if (resp.logFile) {
						outputLink = ` - <a href="/logs/`+resp.logFile+`" target="_blank">View output</a>`
					}
					switch (resp.error) {
						case "not-logged-in":
							alertify.error("You aren't logged in"+outputLink);
							reak;

						case "machine-connection-failed":
							alertify.error("Failed to connect to the remote host machine"+outputLink);
							break;

						default:
							alertify.error("Something went wrong"+outputLink);
							break;	
					}
					if (resp.error != "")
						console.log(resp.error);
					return
				}
			},
			complete: function() {
				setServerRowLoading(this.server, false);
			}
		});
	}
}

function stopServer() {
	if (selectedServers.length == 0 || !$("#stop-server-btn").data("can-click"))
		return;

	$("#stop-server-btn").data("can-click", false);
	setTimeout(function() {
		$("#stop-server-btn").data("can-click", true);
	}, 500);

	for (var i = selectedServers.length-1; i >= 0; i--) {
		var s = $("#"+selectedServers[i]);
		setServerRowLoading(s, true);
		var serverID = s.data("server-id");
		var machineID = s.data("machine-id");
		$.ajax({
			type: "post",
			url: "/server/stop",
			server: s,
			data: {"serverID": serverID, "machineID": machineID},
			dataType: "json",
			async: true,
			success: function(resp) {
				if (!resp.success) {
					var outputLink = "";
					if (resp.logFile) {
						outputLink = ` - <a href="/logs/`+resp.logFile+`" target="_blank">View output</a>`
					}
					switch (resp.error) {
						case "not-logged-in":
							alertify.error("You aren't logged in"+outputLink);
							reak;

						case "machine-connection-failed":
							alertify.error("Failed to connect to the remote host machine"+outputLink);
							break;

						default:
							alertify.error("Something went wrong"+outputLink);
							break;	
					}
					if (resp.error != "")
						console.log(resp.error);
					return
				}
			},
			complete: function() {
				setServerRowLoading(this.server, false);
			}
		});
	}
}

