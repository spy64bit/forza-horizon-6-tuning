export namespace main {
	
	export class CarTune {
	    id: number;
	    name: string;
	    notes: string;
	    updated: number;
	    tirePressureFront?: number;
	    tirePressureRear?: number;
	    finalDrive?: number;
	    gear1?: number;
	    gear2?: number;
	    gear3?: number;
	    gear4?: number;
	    gear5?: number;
	    gear6?: number;
	    gear7?: number;
	    gear8?: number;
	    gear9?: number;
	    gear10?: number;
	    camberFront?: number;
	    camberRear?: number;
	    toeFront?: number;
	    toeRear?: number;
	    casterAngle?: number;
	    antirollFront?: number;
	    antirollRear?: number;
	    springFront?: number;
	    springRear?: number;
	    rideHeightFront?: number;
	    rideHeightRear?: number;
	    reboundFront?: number;
	    reboundRear?: number;
	    bumpFront?: number;
	    bumpRear?: number;
	    downforceFront?: number;
	    downforceRear?: number;
	    brakeBalance?: number;
	    brakePressure?: number;
	    diffFrontAccel?: number;
	    diffFrontDecel?: number;
	    diffRearAccel?: number;
	    diffRearDecel?: number;
	    diffCenter?: number;
	
	    static createFrom(source: any = {}) {
	        return new CarTune(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.notes = source["notes"];
	        this.updated = source["updated"];
	        this.tirePressureFront = source["tirePressureFront"];
	        this.tirePressureRear = source["tirePressureRear"];
	        this.finalDrive = source["finalDrive"];
	        this.gear1 = source["gear1"];
	        this.gear2 = source["gear2"];
	        this.gear3 = source["gear3"];
	        this.gear4 = source["gear4"];
	        this.gear5 = source["gear5"];
	        this.gear6 = source["gear6"];
	        this.gear7 = source["gear7"];
	        this.gear8 = source["gear8"];
	        this.gear9 = source["gear9"];
	        this.gear10 = source["gear10"];
	        this.camberFront = source["camberFront"];
	        this.camberRear = source["camberRear"];
	        this.toeFront = source["toeFront"];
	        this.toeRear = source["toeRear"];
	        this.casterAngle = source["casterAngle"];
	        this.antirollFront = source["antirollFront"];
	        this.antirollRear = source["antirollRear"];
	        this.springFront = source["springFront"];
	        this.springRear = source["springRear"];
	        this.rideHeightFront = source["rideHeightFront"];
	        this.rideHeightRear = source["rideHeightRear"];
	        this.reboundFront = source["reboundFront"];
	        this.reboundRear = source["reboundRear"];
	        this.bumpFront = source["bumpFront"];
	        this.bumpRear = source["bumpRear"];
	        this.downforceFront = source["downforceFront"];
	        this.downforceRear = source["downforceRear"];
	        this.brakeBalance = source["brakeBalance"];
	        this.brakePressure = source["brakePressure"];
	        this.diffFrontAccel = source["diffFrontAccel"];
	        this.diffFrontDecel = source["diffFrontDecel"];
	        this.diffRearAccel = source["diffRearAccel"];
	        this.diffRearDecel = source["diffRearDecel"];
	        this.diffCenter = source["diffCenter"];
	    }
	}

}

