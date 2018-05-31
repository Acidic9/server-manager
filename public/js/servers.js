var machines = {};
var selectedServers = [];
reloadMachines();
$(document).ready(function () {
    $(document).on('change', "input:checkbox.server-checkbox", onServerCheckboxChange);
    var serversWS = new WebSocket("ws://" + window.location.host + "/server/status");
    serversWS.onmessage = onServerStatusMsg;
    $(".select-all-checkbox").click(onSelectAllCheckboxClick);
    $("#start-server-btn").data("can-click", true);
    $("#start-server-btn").click(startServer);
    $("#stop-server-btn").data("can-click", true);
    $("#stop-server-btn").click(stopServer);
});
function reloadMachines() {
    $(".servers-table").each(function () {
        var machineID = $(this).data("machine-id");
        var tag = $(this).prev(".tag");
        var serversTable = $(this);
        tag.hide();
        serversTable.hide();
        var offlineCount = 0;
        var totalServers = 0;
        var isShown = false;
        for (var mID in machines) {
            for (var sID in machines[mID]) {
                if (mID == machineID) {
                    totalServers++;
                    var server = machines[mID][sID];
                    if (server.status == "offline")
                        offlineCount++;
                    if (!isShown) {
                        isShown = true;
                        tag.show();
                        serversTable.show();
                    }
                }
            }
        }
        tag.addClass("is-primary");
        tag.removeClass("is-danger");
        if (offlineCount > 0 && offlineCount == totalServers) {
            tag.removeClass("is-primary");
            tag.addClass("is-danger");
        }
    });
}
function onServerCheckboxChange() {
    var cbox = $(this);
    var tr = cbox.closest("tr.server-row");
    var machineID = tr.data("machine-id");
    var serverID = tr.data("server-id");
    if (cbox.is(":checked")) {
        selectedServers.pushIfNotExist([machineID, serverID]);
        cbox.closest("tr").addClass("is-selected");
    }
    else {
        selectedServers.delete([machineID, serverID]);
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
    var data = JSON.parse(msg.data);
    var server = new Server(data.ID, data.MachineID, data.Game, data.Status, data.Path);
    var machineFound = false;
    var serverFound = false;
    for (var machineID in machines) {
        if (machineID === server.machineID.toString()) {
            machineFound = true;
            for (var serverID in machines[machineID]) {
                if (serverID === server.id.toString()) {
                    serverFound = true;
                    machines[machineID][serverID] = server;
                    break;
                }
            }
        }
    }
    if (!machineFound) {
        machines[server.machineID] = {};
        machines[server.machineID][server.id] = server;
    }
    var status;
    switch (server.status) {
        case "...":
            status = '<span class="tag server-status"><i class="fa fa-refresh spinning" style="color: #4285F4" aria-hidden="true"></i></span>';
            break;
        case "starting":
        case "stopping":
            status = '<span class="tag is-warning server-status"><i class="fa fa-refresh spinning mr4" aria-hidden="true"></i>' + server.status + '</span>';
            break;
        case "online":
        case "running":
            status = '<span class="tag is-success server-status">' + server.status + '</span>';
            break;
        case "offline":
        case "stopped":
            status = '<span class="tag is-danger server-status">' + server.status + '</span>';
            break;
        case "installing":
            status = '<span class="tag is-info server-status"><i class="fa fa-spinner spinning mr4" aria-hidden="true"></i>' + server.status + '</span>';
            break;
        case "":
            status = '<span class="tag is-warning server-status">idle</span>';
            break;
        default:
            status = '<span class="tag is-primary server-status">' + server.status + '</span>';
            break;
    }
    var tr = `<td class="small-column">
				<input type="checkbox" class="server-checkbox">
			</td>
			<td class="small-column server-status-col">` + status + `</td>
			<td>
				<small><i>No Name Set</i></small>
			</td>
			<td class="server-game-col">` + server.game + `</td>
			<td>
				<a title="Manage"><i class="fa fa-wrench" aria-hidden="true" style="padding: 0 10px; margin-top: 2px"></i></a>
			</td>`;
    var found = false;
    $(".server-row").each(function () {
        var serverID = $(this).data("server-id");
        var machineID = $(this).data("machine-id");
        if (server.id == serverID && server.machineID == machineID) {
            found = true;
            var statusText = $(this).find(".server-status-col").text();
            if (statusText != server.status) {
                $(this).find(".server-status-col").html(status);
            }
            var gameText = $(this).find(".server-game-col").text();
            if (gameText != server.game) {
                $(this).find(".server-game-col").html(server.game);
            }
            return false;
        }
    });
    if (!found) {
        $(".servers-table").each(function () {
            var machineID = $(this).data("machine-id");
            if (machineID == server.machineID) {
                $(this).find("tbody").append(`<tr id="server-` + server.id + `-` + server.machineID + `" class="server-row" data-server-id="` + server.id + `" data-machine-id="` + server.machineID + `">` + tr + `</tr>`);
            }
        });
    }
    reloadMachines();
}
function deselectServers() {
    selectedServers = [];
    $(".servers-table tr").each(function () {
        $(this).removeClass("is-selected");
        $(this).find("input:checkbox").prop("checked", false);
    });
}
function setServerRowLoading(tr, loading) {
}
function onSelectAllCheckboxClick() {
    if ($(this).is(":checked")) {
        $("input:checkbox.server-checkbox").prop("checked", true);
    }
    else {
        $("input:checkbox.server-checkbox").prop("checked", false);
    }
    $("input:checkbox.server-checkbox").trigger("change");
}
function startServer() {
    if (selectedServers.length == 0 || !$("#start-server-btn").data("can-click"))
        return;
    $("#start-server-btn").data("can-click", false);
    setTimeout(function () {
        $("#start-server-btn").data("can-click", true);
    }, 500);
    selectedServers.forEach(s => {
        var machineID = s[0];
        var serverID = s[1];
        var server = machines[machineID][serverID];
        setServerRowLoading(server, true);
        server.start()
            .done(resp => {
            if (resp.error)
                console.warn(resp.error);
            if (!resp.success) {
                var outputLink = "";
                if (resp.logFile) {
                    outputLink = ` - <a href="/logs/` + resp.logFile + `" target="_blank">View output</a>`;
                }
                showError(resp.error, true, errorMsg => {
                    return errorMsg + outputLink;
                });
                return;
            }
        })
            .always(function () {
            setServerRowLoading(this.server, false);
        });
    });
}
function stopServer() {
    if (selectedServers.length == 0 || !$("#stop-server-btn").data("can-click"))
        return;
    $("#stop-server-btn").data("can-click", false);
    setTimeout(function () {
        $("#stop-server-btn").data("can-click", true);
    }, 500);
    selectedServers.forEach(s => {
        var machineID = s[0];
        var serverID = s[1];
        var server = machines[machineID][serverID];
        setServerRowLoading(server, true);
        server.stop()
            .done(resp => {
            if (resp.error)
                console.warn(resp.error);
            if (!resp.success) {
                var outputLink = "";
                if (resp.logFile) {
                    outputLink = ` - <a href="/logs/` + resp.logFile + `" target="_blank">View output</a>`;
                }
                showError(resp.error, true, errorMsg => { return errorMsg + outputLink; });
                return;
            }
        })
            .always(function () {
            setServerRowLoading(this.server, false);
        });
    });
}
