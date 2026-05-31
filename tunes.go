package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// CarTune holds a single car tuning setup. All numeric setup fields are
// pointers so an omitted/unknown property is stored as NULL rather than 0.
type CarTune struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Notes   string `json:"notes"`
	Updated int64  `json:"updated"` // unix millis

	// Tires — tire pressure (1.0 - 3.8)
	TirePressureFront *float64 `json:"tirePressureFront"`
	TirePressureRear  *float64 `json:"tirePressureRear"`

	// Gearing
	FinalDrive *float64 `json:"finalDrive"` // 2.2 - 6.1
	Gear1      *float64 `json:"gear1"`      // 0.48 - 6
	Gear2      *float64 `json:"gear2"`
	Gear3      *float64 `json:"gear3"`
	Gear4      *float64 `json:"gear4"`
	Gear5      *float64 `json:"gear5"`
	Gear6      *float64 `json:"gear6"`
	Gear7      *float64 `json:"gear7"`
	Gear8      *float64 `json:"gear8"`
	Gear9      *float64 `json:"gear9"`
	Gear10     *float64 `json:"gear10"`

	// Alignment — camber/toe (-5 - 5), caster (1.0 - 7.0)
	CamberFront *float64 `json:"camberFront"`
	CamberRear  *float64 `json:"camberRear"`
	ToeFront    *float64 `json:"toeFront"`
	ToeRear     *float64 `json:"toeRear"`
	CasterAngle *float64 `json:"casterAngle"`

	// Antiroll bars (1 - 65)
	AntirollFront *float64 `json:"antirollFront"`
	AntirollRear  *float64 `json:"antirollRear"`

	// Springs n/mm (536.5 - 2682.5)
	SpringFront *float64 `json:"springFront"`
	SpringRear  *float64 `json:"springRear"`

	// Ride height cm — front (6 - 11), rear (10.5 - 15.5)
	RideHeightFront *float64 `json:"rideHeightFront"`
	RideHeightRear  *float64 `json:"rideHeightRear"`

	// Damping (1 - 20)
	ReboundFront *float64 `json:"reboundFront"`
	ReboundRear  *float64 `json:"reboundRear"`
	BumpFront    *float64 `json:"bumpFront"`
	BumpRear     *float64 `json:"bumpRear"`

	// Aero downforce kgf — front (105 - 315), rear (117 - 507)
	DownforceFront *float64 `json:"downforceFront"`
	DownforceRear  *float64 `json:"downforceRear"`

	// Brake — balance % (0 - 100), pressure % (0 - 200)
	BrakeBalance  *float64 `json:"brakeBalance"`
	BrakePressure *float64 `json:"brakePressure"`

	// Differential % (0 - 100)
	DiffFrontAccel *float64 `json:"diffFrontAccel"`
	DiffFrontDecel *float64 `json:"diffFrontDecel"`
	DiffRearAccel  *float64 `json:"diffRearAccel"`
	DiffRearDecel  *float64 `json:"diffRearDecel"`
	DiffCenter     *float64 `json:"diffCenter"` // balance 0 - 100
}

var (
	tuneDB   *sql.DB
	tuneDBMu sync.Mutex
)

// tunesDBPath returns (and lazily creates the parent dir for) the tunes database.
func tunesDBPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	dir := filepath.Join(base, "ForzaHorizon6Telemetry")
	os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "tunes.db")
}

// initTunesDB opens the SQLite database and ensures the schema exists.
func initTunesDB() error {
	db, err := sql.Open("sqlite", tunesDBPath())
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1) // SQLite: serialize writes, avoid "database is locked"
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		db.Close()
		return err
	}
	if _, err := db.Exec(tunesSchema); err != nil {
		db.Close()
		return err
	}
	tuneDB = db
	return nil
}

const tunesSchema = `
CREATE TABLE IF NOT EXISTS tunes (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	name             TEXT NOT NULL DEFAULT '',
	notes            TEXT NOT NULL DEFAULT '',
	updated          INTEGER NOT NULL DEFAULT 0,
	tirePressureFront REAL, tirePressureRear REAL,
	finalDrive REAL,
	gear1 REAL, gear2 REAL, gear3 REAL, gear4 REAL, gear5 REAL,
	gear6 REAL, gear7 REAL, gear8 REAL, gear9 REAL, gear10 REAL,
	camberFront REAL, camberRear REAL, toeFront REAL, toeRear REAL, casterAngle REAL,
	antirollFront REAL, antirollRear REAL,
	springFront REAL, springRear REAL,
	rideHeightFront REAL, rideHeightRear REAL,
	reboundFront REAL, reboundRear REAL, bumpFront REAL, bumpRear REAL,
	downforceFront REAL, downforceRear REAL,
	brakeBalance REAL, brakePressure REAL,
	diffFrontAccel REAL, diffFrontDecel REAL,
	diffRearAccel REAL, diffRearDecel REAL, diffCenter REAL
);
CREATE INDEX IF NOT EXISTS idx_tunes_name ON tunes(name);
`

// tuneColumns lists the setup columns in a fixed order shared by read & write.
var tuneColumns = []string{
	"tirePressureFront", "tirePressureRear",
	"finalDrive",
	"gear1", "gear2", "gear3", "gear4", "gear5",
	"gear6", "gear7", "gear8", "gear9", "gear10",
	"camberFront", "camberRear", "toeFront", "toeRear", "casterAngle",
	"antirollFront", "antirollRear",
	"springFront", "springRear",
	"rideHeightFront", "rideHeightRear",
	"reboundFront", "reboundRear", "bumpFront", "bumpRear",
	"downforceFront", "downforceRear",
	"brakeBalance", "brakePressure",
	"diffFrontAccel", "diffFrontDecel",
	"diffRearAccel", "diffRearDecel", "diffCenter",
}

// fieldPtrs returns pointers to the setup fields of t in tuneColumns order.
func (t *CarTune) fieldPtrs() []**float64 {
	return []**float64{
		&t.TirePressureFront, &t.TirePressureRear,
		&t.FinalDrive,
		&t.Gear1, &t.Gear2, &t.Gear3, &t.Gear4, &t.Gear5,
		&t.Gear6, &t.Gear7, &t.Gear8, &t.Gear9, &t.Gear10,
		&t.CamberFront, &t.CamberRear, &t.ToeFront, &t.ToeRear, &t.CasterAngle,
		&t.AntirollFront, &t.AntirollRear,
		&t.SpringFront, &t.SpringRear,
		&t.RideHeightFront, &t.RideHeightRear,
		&t.ReboundFront, &t.ReboundRear, &t.BumpFront, &t.BumpRear,
		&t.DownforceFront, &t.DownforceRear,
		&t.BrakeBalance, &t.BrakePressure,
		&t.DiffFrontAccel, &t.DiffFrontDecel,
		&t.DiffRearAccel, &t.DiffRearDecel, &t.DiffCenter,
	}
}

// values returns the setup field values (as interface{}) in tuneColumns order.
func (t *CarTune) values() []interface{} {
	ptrs := t.fieldPtrs()
	vals := make([]interface{}, len(ptrs))
	for i, p := range ptrs {
		if *p != nil {
			vals[i] = **p
		} else {
			vals[i] = nil
		}
	}
	return vals
}

// scanDest returns scan destinations (sql.NullFloat64) for the setup fields.
func scanDest(n int) []sql.NullFloat64 {
	return make([]sql.NullFloat64, n)
}

func (t *CarTune) applyScan(nf []sql.NullFloat64) {
	ptrs := t.fieldPtrs()
	for i, p := range ptrs {
		if nf[i].Valid {
			v := nf[i].Float64
			*p = &v
		} else {
			*p = nil
		}
	}
}

// ListTunes returns all tunes, optionally filtered by a name substring (case
// insensitive), newest first. Pass "" to return everything.
func (a *App) ListTunes(search string) []CarTune {
	tuneDBMu.Lock()
	defer tuneDBMu.Unlock()
	if tuneDB == nil {
		return []CarTune{}
	}
	cols := "id, name, notes, updated, " + joinCols(tuneColumns)
	query := "SELECT " + cols + " FROM tunes"
	var args []interface{}
	if search != "" {
		query += " WHERE name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	query += " ORDER BY updated DESC, id DESC"

	rows, err := tuneDB.Query(query, args...)
	if err != nil {
		return []CarTune{}
	}
	defer rows.Close()

	out := []CarTune{}
	for rows.Next() {
		var t CarTune
		nf := scanDest(len(tuneColumns))
		dest := make([]interface{}, 0, 4+len(nf))
		dest = append(dest, &t.ID, &t.Name, &t.Notes, &t.Updated)
		for i := range nf {
			dest = append(dest, &nf[i])
		}
		if err := rows.Scan(dest...); err != nil {
			continue
		}
		t.applyScan(nf)
		out = append(out, t)
	}
	return out
}

// SaveTune inserts a new tune (id == 0) or updates an existing one, returning
// the saved tune (with assigned id) or an error message.
func (a *App) SaveTune(t CarTune) (CarTune, string) {
	tuneDBMu.Lock()
	defer tuneDBMu.Unlock()
	if tuneDB == nil {
		return t, "database not initialized"
	}
	t.Updated = time.Now().UnixMilli()

	if t.ID == 0 {
		cols := joinCols(tuneColumns)
		placeholders := "?, ?, ?, " + placeholderList(len(tuneColumns))
		query := "INSERT INTO tunes (name, notes, updated, " + cols + ") VALUES (" + placeholders + ")"
		args := append([]interface{}{t.Name, t.Notes, t.Updated}, t.values()...)
		res, err := tuneDB.Exec(query, args...)
		if err != nil {
			return t, err.Error()
		}
		id, _ := res.LastInsertId()
		t.ID = id
		return t, ""
	}

	setClause := "name = ?, notes = ?, updated = ?"
	for _, c := range tuneColumns {
		setClause += ", " + c + " = ?"
	}
	query := "UPDATE tunes SET " + setClause + " WHERE id = ?"
	args := append([]interface{}{t.Name, t.Notes, t.Updated}, t.values()...)
	args = append(args, t.ID)
	if _, err := tuneDB.Exec(query, args...); err != nil {
		return t, err.Error()
	}
	return t, ""
}

// DeleteTune removes a tune by id, returning an error message or "".
func (a *App) DeleteTune(id int64) string {
	tuneDBMu.Lock()
	defer tuneDBMu.Unlock()
	if tuneDB == nil {
		return "database not initialized"
	}
	if _, err := tuneDB.Exec("DELETE FROM tunes WHERE id = ?", id); err != nil {
		return err.Error()
	}
	return ""
}

// ImportTunes bulk-inserts tunes (all as new rows) inside a single transaction,
// returning the number imported and an error message.
func (a *App) ImportTunes(tunes []CarTune) (int, string) {
	tuneDBMu.Lock()
	defer tuneDBMu.Unlock()
	if tuneDB == nil {
		return 0, "database not initialized"
	}
	tx, err := tuneDB.Begin()
	if err != nil {
		return 0, err.Error()
	}
	cols := joinCols(tuneColumns)
	placeholders := "?, ?, ?, " + placeholderList(len(tuneColumns))
	query := "INSERT INTO tunes (name, notes, updated, " + cols + ") VALUES (" + placeholders + ")"
	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return 0, err.Error()
	}
	defer stmt.Close()

	now := time.Now().UnixMilli()
	count := 0
	for _, t := range tunes {
		updated := t.Updated
		if updated == 0 {
			updated = now
		}
		args := append([]interface{}{t.Name, t.Notes, updated}, t.values()...)
		if _, err := stmt.Exec(args...); err != nil {
			tx.Rollback()
			return 0, err.Error()
		}
		count++
	}
	if err := tx.Commit(); err != nil {
		return 0, err.Error()
	}
	return count, ""
}

func joinCols(cols []string) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += ", "
		}
		out += c
	}
	return out
}

func placeholderList(n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ", "
		}
		out += "?"
	}
	return out
}
