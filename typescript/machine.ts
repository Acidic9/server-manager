class Machine {
	constructor(
		public id: number,
		public title?: string,
		public addr?: string,
		public port?: number,
		public username?: string,
		public password?: string) { }

	delete(async: boolean = true): JQuery.jqXHR<any>  {
		var machine = this;
		var data = { "machineID": this.id };
		return $.ajax({
			type: "post",
			url: "/machine/delete",
			//machine: machine,
			data: data,
			dataType: "json",
			async: async,
		});
	}
}