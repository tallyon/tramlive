package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// VehiclesResponse is response with vehicles data
type VehiclesResponse struct {
	LastUpdate int64         `json:"lastUpdate"`
	Vehicles   []VehicleData `json:"vehicles"`
}

// VehicleData is single object in vehicles array in response
type VehicleData struct {
	ID        string        `json:"id"`
	IsDeleted bool          `json:"isDeleted"`
	Category  string        `json:"category"`
	Color     string        `json:"color"`
	TripID    string        `json:"tripId"`
	Name      string        `json:"name"`
	Path      []VehiclePath `json:"path"`
	Longitude int           `json:"longitude"`
	Latitude  int           `json:"latitude"`
	Heading   int           `json:"heading"`
}

// VehiclePath is path data object specified in path array
type VehiclePath struct {
	Length float64 `json:"length"`
	Y1     int     `json:"y1"`
	Y2     int     `json:"y2"`
	X1     int     `json:"x1"`
	X2     int     `json:"x2"`
	Angle  int     `json:"angle"`
}

// VehicleDataFinal is object that represents final data that is held and sent to requests
type VehicleDataFinal struct {
	Vehicles   []VehicleData
	LastUpdate time.Time
}

// CurrentData is global data object
var CurrentData VehicleDataFinal

func main() {
	CurrentData := &VehicleDataFinal{}
	err := CurrentData.UpdateVehiclesData()
	if err != nil {
		log.Fatal(err)
	}

	// Start periodic update of data
	go CurrentData.StartContinuousDataUpdate()

	http.Handle("/tram", CurrentData)
	http.ListenAndServe(":8080", nil)
}

func (d *VehicleDataFinal) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	dataBytes, err := json.Marshal(d)
	if err != nil {
		res.Write([]byte(""))
		return
	}
	res.Write(dataBytes)
}

// StartContinuousDataUpdate will goroutine every n seconds to update vehicles data
func (d *VehicleDataFinal) StartContinuousDataUpdate() {
	fmt.Printf("\nStarting update routine\n")
	for {
		time.Sleep(5 * time.Second)
		fmt.Printf("Updating... ")
		err := d.UpdateVehiclesData()
		if err != nil {
			fmt.Printf("update failed!\n")
		} else {
			fmt.Printf("update succeded!\n")
		}
	}
}

// UpdateVehiclesData returns last update time and array of all existing vehicles
func (d *VehicleDataFinal) UpdateVehiclesData() error {
	vehicles, err := http.Get("http://www.ttss.krakow.pl/internetservice/geoserviceDispatcher/services/vehicleinfo/vehicles")
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(vehicles.Body)
	vehicles.Body.Close()
	if err != nil {
		return err
	}
	response := VehiclesResponse{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	var existingVehicles []VehicleData
	for _, vehicle := range response.Vehicles {
		if vehicle.IsDeleted == false {
			existingVehicles = append(existingVehicles, vehicle)
		}
	}

	timestamp := time.Unix(response.LastUpdate/1000, 0)
	d.LastUpdate = timestamp
	d.Vehicles = existingVehicles

	return nil
}
