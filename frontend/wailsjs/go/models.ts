export namespace appshell {
	
	export class LocalVaultImportPickDTO {
	    picked: boolean;
	    suggestedName: string;
	
	    static createFrom(source: any = {}) {
	        return new LocalVaultImportPickDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.picked = source["picked"];
	        this.suggestedName = source["suggestedName"];
	    }
	}
	export class ProfileDTO {
	    id: string;
	    displayName: string;
	    repoUrl: string;
	    branch: string;
	    localOnly: boolean;
	    hasPat: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.displayName = source["displayName"];
	        this.repoUrl = source["repoUrl"];
	        this.branch = source["branch"];
	        this.localOnly = source["localOnly"];
	        this.hasPat = source["hasPat"];
	    }
	}
	export class RemoteSaveResult {
	    migrated: boolean;
	    oldRepoUrl: string;
	    newRepoUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new RemoteSaveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.migrated = source["migrated"];
	        this.oldRepoUrl = source["oldRepoUrl"];
	        this.newRepoUrl = source["newRepoUrl"];
	    }
	}
	export class SessionDTO {
	    unlocked: boolean;
	    profileId: string;
	    displayName: string;
	    repoUrl: string;
	    branch: string;
	    dirty: boolean;
	    entryDirty: boolean;
	    pendingSync: boolean;
	    localOnly: boolean;
	    lastError: string;
	    autoLockMinutes: number;
	    lastPullAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.unlocked = source["unlocked"];
	        this.profileId = source["profileId"];
	        this.displayName = source["displayName"];
	        this.repoUrl = source["repoUrl"];
	        this.branch = source["branch"];
	        this.dirty = source["dirty"];
	        this.entryDirty = source["entryDirty"];
	        this.pendingSync = source["pendingSync"];
	        this.localOnly = source["localOnly"];
	        this.lastError = source["lastError"];
	        this.autoLockMinutes = source["autoLockMinutes"];
	        this.lastPullAt = source["lastPullAt"];
	    }
	}

}

export namespace vaultcore {
	
	export class Description {
	    id: string;
	    folderId: string;
	    title: string;
	    key: string;
	    password: string;
	    value: string;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Description(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.folderId = source["folderId"];
	        this.title = source["title"];
	        this.key = source["key"];
	        this.password = source["password"];
	        this.value = source["value"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Folder {
	    id: string;
	    name: string;
	    order: number;
	
	    static createFrom(source: any = {}) {
	        return new Folder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.order = source["order"];
	    }
	}
	export class Vault {
	    version: number;
	    folders: Folder[];
	    descriptions: Description[];
	
	    static createFrom(source: any = {}) {
	        return new Vault(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.folders = this.convertValues(source["folders"], Folder);
	        this.descriptions = this.convertValues(source["descriptions"], Description);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

