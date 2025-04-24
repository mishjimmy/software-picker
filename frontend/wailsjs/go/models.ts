export namespace main {
	
	export class ParadigmVersion {
	    name: string;
	    version: string;
	    path: string;
	    executablePath: string;
	
	    static createFrom(source: any = {}) {
	        return new ParadigmVersion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.path = source["path"];
	        this.executablePath = source["executablePath"];
	    }
	}

}

