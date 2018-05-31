class Machine {
    constructor(id, title, addr, port, username, password) {
        this.id = id;
        this.title = title;
        this.addr = addr;
        this.port = port;
        this.username = username;
        this.password = password;
    }
    delete(async = true) {
        var machine = this;
        var data = { "machineID": this.id };
        return $.ajax({
            type: "post",
            url: "/machine/delete",
            data: data,
            dataType: "json",
            async: async,
        });
    }
}
