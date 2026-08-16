package miniredis

import (
	"testing"

	"github.com/alicebob/miniredis/v2/proto"
)

func TestGeosearch(t *testing.T) {
	_, c := runWithClient(t)

	must1(t, c, "GEOADD", "Sicily", "13.361389", "38.115556", "Palermo")
	must1(t, c, "GEOADD", "Sicily", "15.087269", "37.502669", "Catania")

	t.Run("FROMLONLAT BYRADIUS", func(t *testing.T) {
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km",
			proto.Strings("Palermo", "Catania"),
		)

		// too small radius
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "1", "km",
			proto.Array(),
		)

		// unknown key
		mustDo(t, c,
			"GEOSEARCH", "Capri", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km",
			proto.Array(),
		)
	})

	t.Run("FROMLONLAT BYRADIUS WITHDIST WITHCOORD", func(t *testing.T) {
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "WITHDIST", "WITHCOORD",
			proto.Array(
				proto.Array(
					proto.String("Palermo"),
					proto.String("190.4424"),
					proto.Strings("13.361389", "38.115556"),
				),
				proto.Array(
					proto.String("Catania"),
					proto.String("56.4413"),
					proto.Strings("15.087267", "37.502668"),
				),
			),
		)
	})

	t.Run("ASC DESC COUNT", func(t *testing.T) {
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "ASC",
			proto.Strings("Catania", "Palermo"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "DESC",
			proto.Strings("Palermo", "Catania"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "ASC", "COUNT", "1",
			proto.Strings("Catania"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "ASC", "COUNT", "1", "ANY",
			proto.Strings("Catania"),
		)
	})

	t.Run("FROMMEMBER", func(t *testing.T) {
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMMEMBER", "Catania", "BYRADIUS", "200", "km", "ASC",
			proto.Strings("Catania", "Palermo"),
		)

		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMMEMBER", "nosuch", "BYRADIUS", "200", "km",
			proto.Error("ERR could not decode requested zset member"),
		)
	})

	t.Run("BYBOX", func(t *testing.T) {
		// large box contains both
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYBOX", "1000", "1000", "km", "ASC",
			proto.Strings("Catania", "Palermo"),
		)

		// narrow box (by height) excludes Palermo
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYBOX", "400", "150", "km",
			proto.Strings("Catania"),
		)

		// narrow box (by width) excludes Palermo
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYBOX", "40", "400", "km",
			proto.Strings("Catania"),
		)

		// box distance matches the radius distance (same haversine)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYBOX", "1000", "1000", "km", "ASC", "WITHDIST",
			proto.Array(
				proto.Strings("Catania", "56.4413"),
				proto.Strings("Palermo", "190.4424"),
			),
		)
	})

	t.Run("error cases", func(t *testing.T) {
		// a duplicated FROM* / BY* option is a plain syntax error
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "FROMMEMBER", "Catania", "BYRADIUS", "200", "km",
			proto.Error(msgSyntaxError),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "BYRADIUS", "200", "km", "WITHDIST", "WITHCOORD",
			proto.Error("ERR exactly one of FROMMEMBER or FROMLONLAT can be specified for GEOSEARCH"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "WITHDIST", "WITHCOORD",
			proto.Error("ERR exactly one of BYRADIUS and BYBOX can be specified for GEOSEARCH"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "mm",
			proto.Error(msgUnsupportedUnit),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "abc", "37", "BYRADIUS", "200", "km",
			proto.Error("ERR value is not a valid float"),
		)
		mustDo(t, c,
			"GEOSEARCH", "Sicily", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km", "COUNT", "0",
			proto.Error("ERR COUNT must be > 0"),
		)
	})

	t.Run("wrong type", func(t *testing.T) {
		mustOK(t, c, "SET", "str", "1")
		mustDo(t, c,
			"GEOSEARCH", "str", "FROMLONLAT", "15", "37", "BYRADIUS", "200", "km",
			proto.Error(ErrWrongType.Error()),
		)
	})
}
