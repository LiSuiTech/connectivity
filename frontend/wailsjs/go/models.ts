export namespace types {
	
	export class ConnectResult {
	    success: boolean;
	    message: string;
	    data: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.data = source["data"];
	    }
	}
	export class Server {
	    id: number;
	    remark: string;
	    host: string;
	    port: number;
	    status: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.remark = source["remark"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.status = source["status"];
	        this.type = source["type"];
	    }
	}
	export class ServerClient {
	    id: number;
	    remark: string;
	    host: string;
	    port: number;
	    status: string;
	    type: string;
	    repeatSend: boolean;
	    repeatInterval: number;
	    sendContent: string;
	
	    static createFrom(source: any = {}) {
	        return new ServerClient(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.remark = source["remark"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.status = source["status"];
	        this.type = source["type"];
	        this.repeatSend = source["repeatSend"];
	        this.repeatInterval = source["repeatInterval"];
	        this.sendContent = source["sendContent"];
	    }
	}

}

