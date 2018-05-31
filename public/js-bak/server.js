class Server {
	/**
	 * @param {number} id
	 * @param {number} machineID
	 * @param {string|null} game
	 * @param {string|null} status
	 * @param {string|null} path
	 */
	constructor(id, machineID, game, status, path) {
		this.id = id;
		this.machineID = machineID;
		this.game = game;
		this.status = status;
		this.path = path;
	}

	/**
	 * @return {jqXHR}
	 */
	start() {
		var server = this;
		var data = {"serverID": this.id, "machineID": this.machineID};
		return $.ajax({
			type: "post",
			url: "/server/start",
			server: server,
			data: data,
			dataType: "json",
			async: true,
		});
	}

	/**
	 * @return {jqXHR}
	 */
	stop() {
		var server = this;
		var data = {"serverID": this.id, "machineID": this.machineID};
		return $.ajax({
			type: "post",
			url: "/server/stop",
			server: server,
			data: data,
			dataType: "json",
			async: true,
		});
	}
}