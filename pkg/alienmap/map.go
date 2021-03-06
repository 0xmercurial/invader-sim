package alienmap

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
)

type City struct {
	name string
}

//String method implemented so that City name can be displayed w/ fmt pkg
func (cn *City) String() string {
	return cn.name
}

type Map struct {
	cities      map[string]*City          //all none-destroyed cities
	connections map[City]map[string]*City //connections (north, south, east, est) for all cities (links)
	occupants   map[*City]mapset.Set      //occupants by city
	destroyed   map[string]bool           //cities that have been destroyed
	numAliens   uint64                    //number of Aliens alive
}

type Config struct {
	Filepath  string
	NumAliens uint64
	NumSteps  uint64
}

func (m *Map) InitMap(c *Config) error {
	m.occupants = make(map[*City]mapset.Set)
	m.destroyed = make(map[string]bool)
	m.numAliens = c.NumAliens
	m.ReadMapFile(c.Filepath)
	return nil
}

//Add city adds a pointer to a City struct to Map.cities
func (m *Map) AddCity(city *City) error {
	if m.cities == nil {
		m.cities = make(map[string]*City)
	}
	name := city.name
	m.cities[name] = city
	// Instantiate the map if not yet done
	if m.connections == nil {
		m.connections = make(map[City]map[string]*City)
	}
	// Instantiate the node's connections
	m.connections[*city] = map[string]*City{"north": nil, "west": nil, "south": nil, "east": nil}
	return nil
}

//AddConnection establishes links between two Cities (assumes valid links)
func (m *Map) AddConnection(cityname1 string, cityname2 string, direction string) error {

	c1 := m.cities[cityname1]
	c2 := m.cities[cityname2]

	//Overwites both directions, assumes that input file contain valid connections.
	//If city A is North of city B, city B is South of city A
	switch direction {
	case "north":
		m.connections[*c1]["north"] = c2
		m.connections[*c2]["south"] = c1
	case "south":
		m.connections[*c1]["south"] = c2
		m.connections[*c2]["north"] = c1
	case "east":
		m.connections[*c1]["east"] = c2
		m.connections[*c2]["west"] = c1
	case "west":
		m.connections[*c1]["west"] = c2
		m.connections[*c2]["east"] = c1
	}
	return nil
}

//RemoveCity removes the City from Map.cities and
//deletes it's connections with other Cities
//TODO: better error handling
func (m *Map) RemoveCity(cityname string) error {

	c1, _ := m.cities[cityname]
	// if !ok {
	// 	return nil
	// }

	// Remove the city from all connections
	c1Connections := m.connections[*c1]
	for _, c2 := range c1Connections {
		if c2 != nil {
			c2Connections := m.connections[*c2]
			// Remove c1 from c2's conncetions
			for direction := range c2Connections {
				c2Neighbor := c2Connections[direction]
				// Check if the pointers are the same
				if c2Neighbor == c1 {
					c2Connections[direction] = nil
					break
				}
			}
		}
	}
	m.destroyed[cityname] = true

	// Remove the cities from the list of cities
	delete(m.cities, cityname)
	delete(m.connections, *c1)
	return nil
}

//PrintMap prints the cities along with their neighbors
func (m *Map) PrintMap() {

	// Sort the keys of cityname -> city mapping
	names := make([]string, 0)
	for c := range m.cities {
		names = append(names, c)
	}
	sort.Strings(names)

	for _, n := range names {
		city := m.cities[n]
		connections := m.connections[*city]

		fmt.Fprint(os.Stdout, city)
		fmt.Fprint(os.Stdout, " ")
		directions := []string{"north", "west", "south", "east"}
		for _, direction := range directions {
			neighborCity := connections[direction]
			if neighborCity == nil {
				continue
			}
			fmt.Fprintf(os.Stdout, " %v=%v", direction, neighborCity)
		}
		fmt.Fprintln(os.Stdout)

	}
	fmt.Fprintln(os.Stdout)
}

// ReadMapFile takes in a filepath and constructs a Map from text
func (m *Map) ReadMapFile(filepath string) error {
	// Assumption: city names dont't have spaces
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, line := range lines {
		cityAndConnections := strings.Split(line, " ")
		// Pull out the cityname and its connections
		c1Name := cityAndConnections[0]
		c1Connections := cityAndConnections[1:]

		// Create the city
		c1 := City{c1Name}

		// Easy add if we're dealing with the first city in the map
		if m.cities == nil {
			m.AddCity(&c1)
		} else {
			_, exists := m.cities[c1Name]
			if !exists {
				m.AddCity(&c1)
			}
		}

		for _, con := range c1Connections {
			dirAndName := strings.Split(con, "=")
			direction, c2Name := dirAndName[0], dirAndName[1]
			_, exists := m.cities[c2Name]
			if !exists {
				c2 := City{c2Name}
				m.AddCity(&c2)
			}
			m.AddConnection(c1Name, c2Name, direction)
		}
	}
	return nil
}

// PickRandomCity picks a random city from Map.cities
func (m *Map) PickRandomCity() *City {
	cities := make([]*City, len(m.cities))
	i := 0
	for _, city := range m.cities {
		cities[i] = city
		i++
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	randCityIdx := r.Intn(len(cities))
	return cities[randCityIdx]
}

// PickRandomNeighbor picks a random City from a mapping of directions
// to other Cities
func (m *Map) PickRandomNeighbor(city *City) *City {
	neighborCitiesMap := m.connections[*city]
	neighborCities := make([]*City, 0)
	for _, city := range neighborCitiesMap {
		if city != nil {
			neighborCities = append(neighborCities, city)
		}
	}
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	randNeighborIdx := r.Intn(len(neighborCities))
	return neighborCities[randNeighborIdx]
}

// hasNeighbors checks if given city has viable neighbors.
// Returns false if neighbors are nil
func (m *Map) hasNeighbors(city *City) bool {
	neighborCitiesMap := m.connections[*city]
	if neighborCitiesMap == nil {
		return false
	}

	for _, neighborCity := range neighborCitiesMap {
		if neighborCity != nil {
			return true
		}
	}

	return false
}

// makeAliens makes a slice of int from min to max (inclusive)
// aliens are identified as unique unsigned 64 bit integers
func makeAliens(min, max uint64) []uint64 {
	a := make([]uint64, max-min+1)
	for i := range a {
		a[i] = min + uint64(i)
	}
	return a
}

//CityNotification notfies user when a city has been destroyed
func (m *Map) CityNotification(cityname string, alien1, alien2 interface{}) {
	msg := fmt.Sprintf("%s has been destroyed by alien %v and alien %v!\n",
		cityname,
		alien1,
		alien2,
	)
	fmt.Fprintf(os.Stdout, "%s", msg)
}

// FirstWave represents the first step in the invasion simulation.
// It instantiates a slice of ints (aliens) and begins populating
// Cities randomly. Per the prompt spec, if two aliens end up in a
// City during this phase, that City is destroyed.
func (m *Map) FirstWave(numAliens uint64) error {
	aliens := makeAliens(1, numAliens)
	for _, a := range aliens {
		if len(m.cities) == 0 || m.numAliens == 0 {
			fmt.Fprintf(os.Stdout, "No aliens or cities found")
			break
		}
		randomCity := m.PickRandomCity()
		_, ok := m.occupants[randomCity]
		if ok {
			m.occupants[randomCity].Add(a)
			if m.occupants[randomCity].Cardinality() > 1 {
				aliens := m.occupants[randomCity].ToSlice()
				m.CityNotification(randomCity.name, aliens[0], aliens[1])
				m.numAliens -= 2
			}
			m.RemoveCity(randomCity.name)
			delete(m.occupants, randomCity)
		} else {
			occupantSet := mapset.NewSet()
			occupantSet.Add(a)
			m.occupants[randomCity] = occupantSet
		}
	}
	return nil
}

// MoveAlien moves removes an alien from the origin city's occupants
// and adds it to the destination city's occupants
func (m *Map) MoveAlien(alien interface{}, origin *City, dest *City) {
	m.occupants[dest].Add(alien)
	m.occupants[origin].Remove(alien)
}

// Simulate runs a an alien invasion simulation based on parameters
// passed: numAliens and steps
// Assumes that, at no point, more than two aliens can be in a city
func (m *Map) Simulate(numAliens uint64, steps uint64) error {
	// if no cities/aliens passed in, return nil
	if len(m.cities) == 0 || m.numAliens == 0 {
		return nil
	}
	m.FirstWave(numAliens)
	// starting at 2, firstWave == step 1
	for step := uint64(2); step <= steps; step++ {

		// movedAliens stores aliens already moved in a given step
		movedAliens := make(map[interface{}]bool)

		// retrieve city and accompanying occupants(aliens)
		for city, cityAliens := range m.occupants {
			_, destroyed := m.destroyed[city.name] //check is city is destroyed
			// If those cities have neighbors, we can move the occcupants one step
			if m.hasNeighbors(city) && !destroyed {
				occcupantsToRemove := make([]interface{}, 0)
				// retrieving alien from cityAliens set
				// if we reach this stage, it can be assumed that only 1 alien is present
				for _, cityOccupant := range cityAliens.ToSlice() {
					_, moved := movedAliens[cityOccupant]
					if moved {
						continue
					}
					// pick random, valid neghbor
					randomCity := m.PickRandomNeighbor(city)

					// Update the neighboring city's slice of occupants
					_, ok := m.occupants[randomCity]             // is this city presently being tracked?
					_, destroyed := m.destroyed[randomCity.name] // is this city destroyed?
					if ok {
						// move alien into new city
						m.MoveAlien(cityOccupant, city, randomCity)
						// check if city that alien moved two now has 2 aliens.
						// if so, destroy that city
						if m.occupants[randomCity].Cardinality() > 1 {
							aliens := m.occupants[randomCity].ToSlice()
							m.CityNotification(randomCity.name, aliens[0], aliens[1])
							m.numAliens -= 2
							m.RemoveCity(randomCity.name)
							delete(m.occupants, randomCity)
						}
					} else if !destroyed {
						// assuming we're not tracking city, and it's not destroyed
						// move alien into new city.
						m.occupants[randomCity] = mapset.NewSet()
						m.MoveAlien(cityOccupant, city, randomCity)
					}
					// keep track of which aliens have already moved in a given step
					movedAliens[cityOccupant] = true
				}
				for removedAlien := range occcupantsToRemove {
					cityAliens.Remove(removedAlien)
				}
			} else {
				continue
			}
		}
	}
	fmt.Fprintln(os.Stdout)
	return nil
}

//Simulation creates, runs, and outputs results of an alien invasion simulation
func Simulation(c *Config) error {
	var m Map
	if err := m.InitMap(c); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Initial Map:\n--------------------")
	m.PrintMap()
	if err := m.Simulate(c.NumAliens, c.NumSteps); err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, "Final Map:\n--------------------")
	m.PrintMap()
	return nil
}
