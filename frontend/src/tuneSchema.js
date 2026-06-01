// Shared car-tuning field schema. Used by both the editor (TuningPanel) and the
// telemetry-view tune selector/readout. Fields are optional; valid ranges are
// enforced by the editor.
export const groups = [
    {
        title: 'Tires — Tire Pressure',
        fields: [
            { key: 'tirePressureFront', label: 'Front', min: 1.0, max: 3.8, step: 0.1 },
            { key: 'tirePressureRear', label: 'Rear', min: 1.0, max: 3.8, step: 0.1 },
        ],
    },
    {
        title: 'Gearing',
        fields: [
            { key: 'finalDrive', label: 'Final Drive', min: 2.2, max: 6.1, step: 0.01 },
            { key: 'gear1', label: '1st', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear2', label: '2nd', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear3', label: '3rd', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear4', label: '4th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear5', label: '5th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear6', label: '6th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear7', label: '7th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear8', label: '8th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear9', label: '9th', min: 0.48, max: 6, step: 0.01 },
            { key: 'gear10', label: '10th', min: 0.48, max: 6, step: 0.01 },
        ],
    },
    {
        title: 'Alignment',
        fields: [
            { key: 'camberFront', label: 'Camber Front', min: -5, max: 5, step: 0.1 },
            { key: 'camberRear', label: 'Camber Rear', min: -5, max: 5, step: 0.1 },
            { key: 'toeFront', label: 'Toe Front', min: -5, max: 5, step: 0.1 },
            { key: 'toeRear', label: 'Toe Rear', min: -5, max: 5, step: 0.1 },
            { key: 'casterAngle', label: 'Caster Angle', min: 1.0, max: 7.0, step: 0.1 },
        ],
    },
    {
        title: 'Antiroll Bars',
        fields: [
            { key: 'antirollFront', label: 'Front', min: 1, max: 65, step: 0.1 },
            { key: 'antirollRear', label: 'Rear', min: 1, max: 65, step: 0.1 },
        ],
    },
    {
        title: 'Springs (N/mm)',
        fields: [
            { key: 'springFront', label: 'Front', min: 536.5, max: 2682.5, step: 0.1 },
            { key: 'springRear', label: 'Rear', min: 536.5, max: 2682.5, step: 0.1 },
        ],
    },
    {
        title: 'Ride Height (cm)',
        fields: [
            { key: 'rideHeightFront', label: 'Front', min: 6, max: 11, step: 0.1 },
            { key: 'rideHeightRear', label: 'Rear', min: 7.8, max: 15.5, step: 0.1 },
        ],
    },
    {
        title: 'Damping',
        fields: [
            { key: 'reboundFront', label: 'Rebound Front', min: 1, max: 20, step: 0.1 },
            { key: 'reboundRear', label: 'Rebound Rear', min: 1, max: 20, step: 0.1 },
            { key: 'bumpFront', label: 'Bump Front', min: 1, max: 20, step: 0.1 },
            { key: 'bumpRear', label: 'Bump Rear', min: 1, max: 20, step: 0.1 },
        ],
    },
    {
        title: 'Aero',
        fields: [
            { key: 'downforceFront', label: 'Front', min: 105, max: 315, step: 1 },
            { key: 'downforceRear', label: 'Rear', min: 117, max: 507, step: 1 },
        ],
    },
    {
        title: 'Brake',
        fields: [
            { key: 'brakeBalance', label: 'Balance (%)', min: 0, max: 100, step: 1 },
            { key: 'brakePressure', label: 'Pressure (%)', min: 0, max: 200, step: 1 },
        ],
    },
    {
        title: 'Differential',
        fields: [
            { key: 'diffFrontAccel', label: 'Front Accel (%)', min: 0, max: 100, step: 1 },
            { key: 'diffFrontDecel', label: 'Front Decel (%)', min: 0, max: 100, step: 1 },
            { key: 'diffRearAccel', label: 'Rear Accel (%)', min: 0, max: 100, step: 1 },
            { key: 'diffRearDecel', label: 'Rear Decel (%)', min: 0, max: 100, step: 1 },
            { key: 'diffCenter', label: 'Center Balance (%)', min: 0, max: 100, step: 1 },
        ],
    },
]

export const allFields = groups.flatMap((g) => g.fields)
