package gps

import (
	"bufio"
	"embed"
	"html/template"
	"log"
	"math"
	"net/http"

	"github.com/merliot/dean"
	"github.com/merliot/sw-poc/models/common"
	"github.com/merliot/sw-poc/models/gps/nmea"
	"github.com/tarm/serial"
)

//go:embed css js template
var fs embed.FS

var indexTmpl = template.Must(template.ParseFS(fs, "template/index.html"))
var buildTmpl = template.Must(template.ParseFS(fs, "template/build.tmpl"))
var deployTmpl = template.Must(template.ParseFS(fs, "template/deploy.tmpl"))

type Gps struct {
	*common.Common
	Lat  float64
	Long float64
}

type Update struct {
	Path string
	Lat  float64
	Long float64
}

func New(id, model, name string) dean.Thinger {
	println("NEW GPS")
	return &Gps{
		Common: common.New(id, model, name).(*common.Common),
	}
}

func (g *Gps) save(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) getState(msg *dean.Msg) {
	g.Path = "state"
	msg.Marshal(g).Reply()
}

func (g *Gps) update(msg *dean.Msg) {
	msg.Unmarshal(g).Broadcast()
}

func (g *Gps) Subscribers() dean.Subscribers {
	return dean.Subscribers{
		"state":     g.save,
		"get/state": g.getState,
		"update":    g.update,
	}
}

func (p *Ps30m) api(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api\n"))
	w.Write([]byte("/deploy?target={target}\n"))
}

func (g *Gps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "":
		g.Index(indexTmpl, w, r)
	case "api":
		g.api(w, r)
	case "deploy.html":
		g.ShowDeploy(deployTmpl, w, r)
	case "deploy":
		g.Deploy(buildTmpl, w, r)
	default:
		g.Common.API(fs, w, r)
	}
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//
//	a given longitude and latitude relatively accurately (using a spherical
//	approximation of the Earth) through the Haversin Distance Formula for
//	great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func (g *Gps) Run(i *dean.Injector) {
	var msg dean.Msg
	var update = Update{Path: "update"}

	cfg := &serial.Config{Name: "/dev/ttyS0", Baud: 9600}
	ser, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(ser)
	for scanner.Scan() {
		lat, long, err := nmea.ParseGLL(scanner.Text())
		if err != nil {
			continue
		}
		dist := int(distance(lat, long, g.Lat, g.Long) * 100.0) // cm
		if dist < 200 /*cm*/ {
			continue
		}
		g.Lat, g.Long = lat, long
		update.Lat, update.Long = lat, long
		i.Inject(msg.Marshal(update))
	}
}
