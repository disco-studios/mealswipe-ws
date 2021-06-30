package validators

import "testing"

func TestLatLonWithinUnitedStates(t *testing.T) {
	t.Run("Hoboken, NJ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(40.74525, -74.034775) == false {
			t.FailNow()
		}
	})
	t.Run("Port Hueneme, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(34.155834, -119.202789) == false {
			t.FailNow()
		}
	})
	t.Run("Auburn, NY, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.93333, -76.566666) == false {
			t.FailNow()
		}
	})
	t.Run("Jamestown, NY, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.09555, -79.238609) == false {
			t.FailNow()
		}
	})
	t.Run("Fulton, MO, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(38.84666, -91.948059) == false {
			t.FailNow()
		}
	})
	t.Run("Bedford, OH, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.39250, -81.534447) == false {
			t.FailNow()
		}
	})
	t.Run("Stuart, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(27.19222, -80.243057) == false {
			t.FailNow()
		}
	})
	t.Run("San Angelo, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			31.442778, -100.450279) == false {
			t.FailNow()
		}
	})
	t.Run("Woodbridge, NJ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(40.56000, -74.290001) == false {
			t.FailNow()
		}
	})
	t.Run("Vista, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.193611, -117.241112) == false {
			t.FailNow()
		}
	})
	t.Run("South Bend, IN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.67638, -86.250275) == false {
			t.FailNow()
		}
	})
	t.Run("Davenport, IA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.54305, -90.590836) == false {
			t.FailNow()
		}
	})
	t.Run("Sparks, NV, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			39.554443, -119.735558) == false {
			t.FailNow()
		}
	})
	t.Run("Green Bay, WI, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(44.51333, -88.015831) == false {
			t.FailNow()
		}
	})
	t.Run("San Mateo, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			37.554169, -122.313057) == false {
			t.FailNow()
		}
	})
	t.Run("Tyler, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.34999, -95.300003) == false {
			t.FailNow()
		}
	})
	t.Run("League City, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(29.49972, -95.089722) == false {
			t.FailNow()
		}
	})
	t.Run("Lewisville, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(33.03833, -97.006111) == false {
			t.FailNow()
		}
	})
	t.Run("Meridian, ID, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			43.614166, -116.398888) == false {
			t.FailNow()
		}
	})
	t.Run("Waterbury, CT, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.55611, -73.041389) == false {
			t.FailNow()
		}
	})
	t.Run("Jurupa Valley, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			34.000000, -117.483330) == false {
			t.FailNow()
		}
	})
	t.Run("West Palm Beach, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(26.70972, -80.064163) == false {
			t.FailNow()
		}
	})
	t.Run("Antioch, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			38.005001, -121.805832) == false {
			t.FailNow()
		}
	})
	t.Run("High Point, NC, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(35.97055, -79.997498) == false {
			t.FailNow()
		}
	})
	t.Run("Miami Gardens, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(25.94212, -80.269920) == false {
			t.FailNow()
		}
	})
	t.Run("Murrieta, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.569443, -117.202499) == false {
			t.FailNow()
		}
	})
	t.Run("Springfield, IL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(39.79999, -89.650002) == false {
			t.FailNow()
		}
	})
	t.Run("El Monte, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			34.073334, -118.027496) == false {
			t.FailNow()
		}
	})
	t.Run("West Jordan, UT, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			40.606388, -111.976112) == false {
			t.FailNow()
		}
	})
	t.Run("College Station, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(30.60138, -96.314445) == false {
			t.FailNow()
		}
	})
	t.Run("Fairfield, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			38.257778, -122.054169) == false {
			t.FailNow()
		}
	})
	t.Run("Evansville, IN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(37.97722, -87.550552) == false {
			t.FailNow()
		}
	})
	t.Run("Cambridge, MA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.37361, -71.110558) == false {
			t.FailNow()
		}
	})
	t.Run("Richardson, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.96555, -96.715836) == false {
			t.FailNow()
		}
	})
	t.Run("Berkeley, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			37.871666, -122.272781) == false {
			t.FailNow()
		}
	})
	t.Run("Columbia, MS, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(38.95156, -92.328636) == false {
			t.FailNow()
		}
	})
	t.Run("Athens, GA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(33.95000, -83.383331) == false {
			t.FailNow()
		}
	})
	t.Run("Lafayette, LA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(30.21666, -92.033333) == false {
			t.FailNow()
		}
	})
	t.Run("Sterling Heights, MI, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.58027, -83.030281) == false {
			t.FailNow()
		}
	})
	t.Run("Visalia, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			36.316666, -119.300003) == false {
			t.FailNow()
		}
	})
	t.Run("Hampton, VA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(37.03494, -76.360123) == false {
			t.FailNow()
		}
	})
	t.Run("West Valley City, UT, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			40.689167, -111.993889) == false {
			t.FailNow()
		}
	})
	t.Run("Surprise, AZ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.630554, -112.366669) == false {
			t.FailNow()
		}
	})
	t.Run("Thornton, CO, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			39.903057, -104.954445) == false {
			t.FailNow()
		}
	})
	t.Run("Miramar, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(25.97888, -80.282501) == false {
			t.FailNow()
		}
	})
	t.Run("Murfreesboro, TN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(35.84611, -86.391945) == false {
			t.FailNow()
		}
	})
	t.Run("Pasadena, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			34.156113, -118.131943) == false {
			t.FailNow()
		}
	})
	t.Run("Bridgeport, CT, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.18639, -73.195557) == false {
			t.FailNow()
		}
	})
	t.Run("Paterson, NJ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(40.91474, -74.162827) == false {
			t.FailNow()
		}
	})
	t.Run("Rockford, Il, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.25944, -89.064445) == false {
			t.FailNow()
		}
	})
	t.Run("Joliet, Illinois, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.52055, -88.150558) == false {
			t.FailNow()
		}
	})
	t.Run("Escondido, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.124722, -117.080833) == false {
			t.FailNow()
		}
	})
	t.Run("Kansas City, KS, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(39.10666, -94.676392) == false {
			t.FailNow()
		}
	})
	t.Run("Springfield, MA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.10139, -72.590279) == false {
			t.FailNow()
		}
	})
	t.Run("Springfield, MO, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(37.21038, -93.297256) == false {
			t.FailNow()
		}
	})
	t.Run("Corona, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.866669, -117.566666) == false {
			t.FailNow()
		}
	})
	t.Run("Pembroke Pines, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(26.01250, -80.313614) == false {
			t.FailNow()
		}
	})
	t.Run("Elk Grove, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			38.438332, -121.381943) == false {
			t.FailNow()
		}
	})
	t.Run("Oceanside, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.211666, -117.325836) == false {
			t.FailNow()
		}
	})
	t.Run("Newport News, VA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(37.07083, -76.484444) == false {
			t.FailNow()
		}
	})
	t.Run("Sioux Falls, South Dakota, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(43.53638, -96.731667) == false {
			t.FailNow()
		}
	})
	t.Run("Vancouver, WA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			45.633331, -122.599998) == false {
			t.FailNow()
		}
	})
	t.Run("Worcester, MA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(42.27138, -71.798889) == false {
			t.FailNow()
		}
	})
	t.Run("Tallahassee, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(30.45500, -84.253334) == false {
			t.FailNow()
		}
	})
	t.Run("Columbus, GA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.49222, -84.940277) == false {
			t.FailNow()
		}
	})
	t.Run("Augusta, GA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(33.46666, -81.966667) == false {
			t.FailNow()
		}
	})
	t.Run("Montgomery, AL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.36166, -86.279167) == false {
			t.FailNow()
		}
	})
	t.Run("Aurora, IL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(41.76388, -88.290001) == false {
			t.FailNow()
		}
	})
	t.Run("Amarillo, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			35.199165, -101.845276) == false {
			t.FailNow()
		}
	})
	t.Run("Modesto, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			37.661388, -120.994446) == false {
			t.FailNow()
		}
	})
	t.Run("Garland, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.90722, -96.635277) == false {
			t.FailNow()
		}
	})
	t.Run("Irvine, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.669445, -117.823059) == false {
			t.FailNow()
		}
	})
	t.Run("Aurora, CO, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			39.710835, -104.812500) == false {
			t.FailNow()
		}
	})
	t.Run("Arlington, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.70500, -97.122780) == false {
			t.FailNow()
		}
	})
	t.Run("Kansas City, MO, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(39.09972, -94.578331) == false {
			t.FailNow()
		}
	})
	t.Run("Memphis, TN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(35.11750, -89.971107) == false {
			t.FailNow()
		}
	})
	t.Run("Indianapolis, IN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(39.79100, -86.148003) == false {
			t.FailNow()
		}
	})
	t.Run("Columbus, OH, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(39.98333, -82.983330) == false {
			t.FailNow()
		}
	})
	t.Run("Austin, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(30.26666, -97.733330) == false {
			t.FailNow()
		}
	})
	t.Run("Dallas, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.77916, -96.808891) == false {
			t.FailNow()
		}
	})
	t.Run("Redwood City, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			37.487846, -122.236115) == false {
			t.FailNow()
		}
	})
	t.Run("Gastonia, NC, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			35.25528, -81.180275) == false {
			t.FailNow()
		}
	})
	t.Run("New Braunfels, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			29.70000, -98.116669) == false {
			t.FailNow()
		}
	})
	t.Run("Palm Beach Gardens, FL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(26.83861, -80.129967) == false {
			t.FailNow()
		}
	})
	t.Run("Forestville, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			38.473625, -122.889992) == false {
			t.FailNow()
		}
	})
	t.Run("Houston, TX, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(29.74990, -95.358421) == false {
			t.FailNow()
		}
	})
	t.Run("Muncie, IN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(40.19189, -85.401695) == false {
			t.FailNow()
		}
	})
	t.Run("Palm Springs, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.830517, -116.545601) == false {
			t.FailNow()
		}
	})
	t.Run("Hot Springs, AR, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(34.49621, -93.057220) == false {
			t.FailNow()
		}
	})
	t.Run("Richmond, VA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(37.54129, -77.434769) == false {
			t.FailNow()
		}
	})
	t.Run("Fayetteville, AR, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(36.08215, -94.171852) == false {
			t.FailNow()
		}
	})
	t.Run("Yuma, AZ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			32.698437, -114.650398) == false {
			t.FailNow()
		}
	})
	t.Run("Peoria, AZ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.580944, -112.237068) == false {
			t.FailNow()
		}
	})
	t.Run("Tempe, AZ, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			33.427204, -111.939896) == false {
			t.FailNow()
		}
	})
	t.Run("Diamond Bar, CA, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(
			34.028622, -117.810333) == false {
			t.FailNow()
		}
	})
	t.Run("Auburn, AL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(32.60985, -85.480782) == false {
			t.FailNow()
		}
	})
	t.Run("Hoover, AL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(33.40574, -86.811546) == false {
			t.FailNow()
		}
	})
	t.Run("Decatur, AL, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(34.60381, -86.985039) == false {
			t.FailNow()
		}
	})
	t.Run("Bloomington, MN, USA valid", func(t *testing.T) {
		if LatLonWithinUnitedStates(44.84079, -93.298279) == false {
			t.FailNow()
		}
	})
	t.Run("Mexico City, MX invalid", func(t *testing.T) {
		if LatLonWithinUnitedStates(19.4326, 99.1332) == true {
			t.FailNow()
		}
	})
	// TODO Alaska
	// TODO More non-us places
}
