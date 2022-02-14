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
	connections map[City]map[string]*City //connections (north, south, east, est) for all cities
	occupants   map[*City]mapset.Set      //occupants by city
	destroyed   map[string]bool           //cities that have been destroyed
	numAliens   uint64
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
	fmt.Fprintln(os.Stdout, "aliens", m.numAliens)
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
	fmt.Fprintln(os.Stdout, cityname, "removed")
	return nil
}

//PrintMap prints the cities along with their neighbors
func (m *Map) PrintMap() {

	fmt.Fprintln(os.Stdout, "cities", m.cities)
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
		for direction, neighborCity := range connections {
			fmt.Fprintf(os.Stdout, " %v=%v", direction, neighborCity)
		}
		fmt.Fprintln(os.Stdout)
	}
}

// ReadMapFile takes in a filepath and constructs a Map from text
func (m *Map) ReadMapFile(filepath string) error {
	// Assumption: city names dont't have spaces
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
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

// FirstWave represents the first step in the invasion simulation.
// It instantiates a slice of ints (aliens) and begins populating
// Cities randomly. Per the prompt spec, if two aliens end up in a
// City during this phase, that City is destroyed.
func (m *Map) FirstWave(numAliens uint64) error {
	fmt.Fprintf(os.Stdout, "first wave\n")
	aliens := makeAliens(1, numAliens)
	fmt.Fprintf(os.Stdout, "made aliens\n")
	for _, a := range aliens {
		if len(m.cities) == 0 || m.numAliens == 0 {
			fmt.Fprintf(os.Stdout, "break")
			break
		}
		randomCity := m.PickRandomCity()
		_, ok := m.occupants[randomCity]
		if ok {
			m.occupants[randomCity].Add(a)
			if m.occupants[randomCity].Cardinality() > 1 {
				aliens := m.occupants[randomCity].ToSlice()
				msg := fmt.Sprintf("%s has been destroyed by alien %d and alien %d!\n",
					randomCity.name,
					aliens[0],
					aliens[1],
				)
				fmt.Fprintf(os.Stdout, "%s", msg)
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

func (m *Map) SubsequentWave() {} //?