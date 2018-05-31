class Server {
	constructor(
		public id: number,
		public machineID: number,
		public game: string,
		public status: string,
		public path: string) { }

	start(async: boolean = true): JQuery.jqXHR<any> {
		var server = this;
		var data = { "serverID": this.id, "machineID": this.machineID };
		return $.ajax({
			type: "post",
			url: "/server/start",
		//	server: server,
			data: data,
			dataType: "json",
			async: async,
		});
	}

	stop(async: boolean = true): JQuery.jqXHR<any>  {
		var server = this;
		var data = { "serverID": this.id, "machineID": this.machineID };
		return $.ajax({
			type: "post",
			url: "/server/stop",
		//	server: server,
			data: data,
			dataType: "json",
			async: async,
		});
	}
}