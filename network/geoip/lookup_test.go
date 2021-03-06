package geoip

import (
	"net"
	"testing"
)

func TestLocationLookup(t *testing.T) {
	ip1 := net.ParseIP("81.2.69.142")
	loc1, err := GetLocation(ip1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", loc1)

	ip2 := net.ParseIP("1.1.1.1")
	loc2, err := GetLocation(ip2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", loc2)

	ip3 := net.ParseIP("8.8.8.8")
	loc3, err := GetLocation(ip3)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", loc3)

	ip4 := net.ParseIP("81.2.70.142")
	loc4, err := GetLocation(ip4)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", loc1)

	dist1 := loc1.EstimateNetworkProximity(loc2)
	dist2 := loc2.EstimateNetworkProximity(loc3)
	dist3 := loc1.EstimateNetworkProximity(loc3)
	dist4 := loc1.EstimateNetworkProximity(loc4)

	t.Logf("proximity %s <> %s: %d", ip1, ip2, dist1)
	t.Logf("proximity %s <> %s: %d", ip2, ip3, dist2)
	t.Logf("proximity %s <> %s: %d", ip1, ip3, dist3)
	t.Logf("proximity %s <> %s: %d", ip1, ip4, dist4)

}
