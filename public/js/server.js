class Server {
    constructor(id, machineID, game, status, path) {
        this.id = id;
        this.machineID = machineID;
        this.game = game;
        this.status = status;
        this.path = path;
    }
    start(async = true) {
        var server = this;
        var data = { "serverID": this.id, "machineID": this.machineID };
        return $.ajax({
            type: "post",
            url: "/server/start",
            data: data,
            dataType: "json",
            async: async,
        });
    }
    stop(async = true) {
        var server = this;
        var data = { "serverID": this.id, "machineID": this.machineID };
        return $.ajax({
            type: "post",
            url: "/server/stop",
            data: data,
            dataType: "json",
            async: async,
        });
    }
}
