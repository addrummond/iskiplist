package iskiplist

/*
The optimal value of p for a general purpose skiplist is is approximately 1/e.
See https://github.com/sean-public/fast-skiplist and the following paper
that it references:
https://www.sciencedirect.com/science/article/pii/030439759400296U

The following Python 3 function can be pasted into the repl. Call
table(n, length) to generate a table.

for _ in range(1): # dummy loop to allow pasting into repl in one go
    from math import *
    def table(n_elems, table_length):
        tot = 0
        for i in range(table_length):
            # if we're tossing n_elems coins, then we need to get exactly i
            # heads at least once, and not ever get more than i heads.
            p_right_number = pow(1/e, i) # probability of at least i heads for single toss sequence
            p_more = pow(1/e, i+1) # probability of at least i+1 heads for single toss sequence
            p = (1 - pow(1-p_right_number, n_elems)) - (1 - pow(1-p_more, n_elems))
            v = max(0, min(1 << 32, round(p * (1 << 32))))
            tot = max(0, min(1 << 32, tot+v))
            print(str(tot) + ",")

We can simulate up to 21 "coin tosses" (where heads has probability 1/e) using
a single unsigned 32-bit random number and a lookup table. If the random number
is >= the last value in the table, then a second random number has to be
generated. When estimating the number of levels for a list of a given length,
we don't need to bother re-rolling in the cases where the probabilities get too
small to be represented by a 32-bit unsigned int. This just means that we don't
very very very rarely assign 30 levels to a short skip list.
*/

func nTosses(l *ISkipList) int {
	// The PCG state has to be odd, so we know that it's uninitialized if the
	// state is zero.
	if l.rand.IsUninitialized() {
		fastSeed(l)
	}

	// Note that a binary search isn't the way to go here, since the value is
	// far more likely to be < one of the first few elements of pTable. A linear
	// search probably isn't quite the probabilistically optimal algorithm, but
	// it's simple and close enough.

	r := l.rand.Random()
	for i := 0; i < len(pTable); i++ {
		if r < pTable[i] {
			return int(i)
		}
	}
	r = l.rand.Random()
	for i := 0; i+len(pTable) < maxLevels; i++ {
		if r < pTable[i] {
			return i + len(pTable)
		}
	}
	return maxLevels
}

func estimateNLevelsFromLength(l *ISkipList, ni int) int {
	// We want the code to handle lengths greater than 2^31, but also to build
	// on i386. In the latter case, 'int' is 32 bits and some of the constants
	// below overflow it. Explicitly casting to a 64-bit int allows the code
	// below to work on both 32-bit and 64-bit architectures.
	n := int64(ni)

	nLevels := 0
outer:
	for n > 0 {
		if n < 8 {
			for ; n >= 0; n-- {
				nt := nTosses(l)
				if nt > nLevels {
					nLevels = nt
				}
			}
			break
		}

		r := l.rand.Random()
		if n < 32 {
			n -= 8
			for i, p := range pTable8 {
				if r < p {
					if i+pTable8Zoff > nLevels {
						nLevels = i + pTable8Zoff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 128 {
			n -= 32
			for i, p := range pTable32 {
				if r < p {
					if i+pTable32ZOff > nLevels {
						nLevels = i + pTable32ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 512 {
			n -= 128
			for i, p := range pTable128 {
				if r < p {
					if i+pTable128ZOff > nLevels {
						nLevels = i + pTable128ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 2048 {
			n -= 512
			for i, p := range pTable512 {
				if r < p {
					if i+pTable512ZOff > nLevels {
						nLevels = i + pTable512ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 8192 {
			n -= 2048
			for i, p := range pTable2048 {
				if r < p {
					if i+pTable2048ZOff > nLevels {
						nLevels = i + pTable2048ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 32768 {
			n -= 8192
			for i, p := range pTable8192 {
				if r < p {
					if i+pTable8192ZOff > nLevels {
						nLevels = i + pTable8192ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 131072 {
			n -= 32768
			for i, p := range pTable32768 {
				if r < p {
					if i+pTable32768ZOff > nLevels {
						nLevels = i + pTable32768ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 262144 {
			n -= 131072
			for i, p := range pTable131072 {
				if r < p {
					if i+pTable131072ZOff > nLevels {
						nLevels = i + pTable131072ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 1048576 {
			n -= 262144
			for i, p := range pTable262144 {
				if r < p {
					if i+pTable262144ZOff > nLevels {
						nLevels = i + pTable262144ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 4194304 {
			n -= 1048576
			for i, p := range pTable1048576 {
				if r < p {
					if i+pTable1048576ZOff > nLevels {
						nLevels = i + pTable1048576ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 16777216 {
			n -= 4194304
			for i, p := range pTable4194304 {
				if r < p {
					if i+pTable4194304ZOff > nLevels {
						nLevels = i + pTable4194304ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 67108864 {
			n -= 16777216
			for i, p := range pTable16777216 {
				if r < p {
					if i+pTable16777216ZOff > nLevels {
						nLevels = i + pTable16777216ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 268435456 {
			n -= 67108864
			for i, p := range pTable67108864 {
				if r < p {
					if i+pTable67108864ZOff > nLevels {
						nLevels = i + pTable67108864ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 1073741824 {
			n -= 268435456
			for i, p := range pTable268435456 {
				if r < p {
					if i+pTable268435456ZOff > nLevels {
						nLevels = i + pTable268435456ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 4294967296 {
			n -= 1073741824
			for i, p := range pTable1073741824 {
				if r < p {
					if i+pTable1073741824ZOff > nLevels {
						nLevels = i + pTable1073741824ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 17179869184 {
			n -= 4294967296
			for i, p := range pTable4294967296 {
				if r < p {
					if i+pTable4294967296ZOff > nLevels {
						nLevels = i + pTable4294967296ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 68719476736 {
			n -= 17179869184
			for i, p := range pTable17179869184 {
				if r < p {
					if i+pTable17179869184ZOff > nLevels {
						nLevels = i + pTable17179869184ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 274877906944 {
			n -= 68719476736
			for i, p := range pTable68719476736 {
				if r < p {
					if i+pTable68719476736ZOff > nLevels {
						nLevels = i + pTable68719476736ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 1099511627776 {
			n -= 274877906944
			for i, p := range pTable274877906944 {
				if r < p {
					if i+pTable274877906944ZOff > nLevels {
						nLevels = i + pTable274877906944ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else if n < 4398046511104 {
			n -= 1099511627776
			for i, p := range pTable1099511627776 {
				if r < p {
					if i+pTable1099511627776ZOff > nLevels {
						nLevels = i + pTable1099511627776ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		} else {
			n -= 4398046511104
			for i, p := range pTable4398046511104 {
				if r < p {
					if i+pTable4398046511104ZOff > nLevels {
						nLevels = i + pTable4398046511104ZOff
					}
					continue outer
				}
			}
			nLevels = maxLevels
			break
		}
	}

	return nLevels
}

var pTable = [...]uint32{
	2714937127,
	3713706680,
	4081133465,
	4216302225,
	4266028033,
	4284321135,
	4291050791,
	4293526493,
	4294437253,
	4294772303,
	4294895561,
	4294940905,
	4294957586,
	4294963723,
	4294965981,
	4294966812,
	4294967118,
	4294967230,
	4294967271,
	4294967286,
	4294967292,
	// nTosses will re-roll if we get here
}

const pTable8Zoff = 0

var pTable8 = [...]uint32{
	109486150,
	1341966772,
	2854482294,
	3704544712,
	4068839996,
	4210533266,
	4263735088,
	4283454413,
	4290728799,
	4293407615,
	4294393464,
	4294756188,
	4294889633,
	4294938726,
	4294956786,
	4294963430,
	4294965874,
	4294966773,
	4294967104,
	4294967226,
	4294967271,
	4294967287,
	4294967293,
	4294967295,
	// EstimateNLevelsFromLength won't estimate more than 24 levels for an
	// ISkipList of length 8-32, since these all have the same approximate
	// probability in the uint32 representation. (Not really any point in
	// re-rolling, since we wouldn't actually want that many levels for such a
	// short list.)
}

const pTable32ZOff = 0

var pTable32 = [...]uint32{
	1814,
	40934310,
	837972623,
	2377167476,
	3459416588,
	3967060536,
	4171394555,
	4249100596,
	4278038387,
	4288731967,
	4292672427,
	4294122923,
	4294656650,
	4294853013,
	4294925253,
	4294951829,
	4294961606,
	4294965203,
	4294966526,
	4294967013,
	4294967192,
	4294967258,
	4294967282,
	4294967291,
	4294967294,
	4294967295,
}

const pTable128ZOff = 1

var pTable128 = [...]uint32{
	35,
	6223572,
	403050565,
	1807722944,
	3126048609,
	3821602355,
	4114418537,
	4227650965,
	4270080238,
	4285795170,
	4291590797,
	4293724845,
	4294510182,
	4294799127,
	4294905429,
	4294944536,
	4294958923,
	4294964216,
	4294966163,
	4294966879,
	4294967143,
	4294967240,
	4294967276,
	4294967289,
	4294967294,
}

const pTable512ZOff = 3

var pTable512 = [...]uint32{
	333088,
	134786966,
	1205322667,
	2692169500,
	3617048133,
	4031966501,
	4196280970,
	4258396152,
	4281477221,
	4289999651,
	4293139136,
	4294294665,
	4294719838,
	4294876261,
	4294933807,
	4294954977,
	4294962765,
	4294965630,
	4294966684,
	4294967072,
	4294967215,
	4294967267,
	4294967286,
	4294967293,
}

const pTable2048ZOff = 4

var pTable2048 = [...]uint32{
	4166,
	26639969,
	663025169,
	2160416911,
	3335708383,
	3913619990,
	4150540523,
	4241260681,
	4275131154,
	4287659314,
	4292277394,
	4293977541,
	4294603160,
	4294833335,
	4294918015,
	4294949167,
	4294960627,
	4294964843,
	4294966394,
	4294966965,
	4294967175,
	4294967252,
	4294967280,
	4294967290,
	4294967294,
	4294967295,
}

const pTable8192ZOff = 5

var pTable8192 = [...]uint32{
	6,
	2439161,
	274960740,
	1562689599,
	2960976977,
	3745752254,
	4084136811,
	4216170712,
	4265809891,
	4284217790,
	4291009643,
	4293510934,
	4294431474,
	4294770171,
	4294894778,
	4294940619,
	4294957483,
	4294963687,
	4294965969,
	4294966809,
	4294967118,
	4294967232,
	4294967274,
	4294967289,
	4294967295,
}

const pTable32768ZOff = 7

var pTable32768 = [...]uint32{
	72144,
	75268160,
	970198364,
	2484715184,
	3511733436,
	3988349089,
	4179519954,
	4252130417,
	4279158542,
	4289144801,
	4292824401,
	4294178844,
	4294677223,
	4294860581,
	4294928037,
	4294952853,
	4294961982,
	4294965340,
	4294966575,
	4294967030,
	4294967197,
	4294967259,
	4294967282,
}

const pTable131072ZOff = 8

var pTable131072 = [...]uint32{
	405,
	11183109,
	481090601,
	1919581637,
	3193692305,
	3851465646,
	4126166244,
	4232080557,
	4271724635,
	4286402130,
	4291814358,
	4293807125,
	4294540456,
	4294810265,
	4294909526,
	4294946043,
	4294959477,
	4294964419,
	4294966237,
	4294966906,
	4294967152,
	4294967242,
}

const pTable262144ZOff = 9

var pTable262144 = [...]uint32{
	29118,
	53888225,
	857932880,
	2374795857,
	3453760321,
	3963999422,
	4170114604,
	4248607756,
	4277854048,
	4288663739,
	4292647271,
	4294113661,
	4294653242,
	4294851760,
	4294924793,
	4294951661,
	4294961545,
	4294965181,
	4294966519,
	4294967011,
	4294967192,
}

const pTable1048576ZOff = 10

var pTable1048576 = [...]uint32{
	106,
	6838072,
	401444533,
	1795929150,
	3116411128,
	3816914031,
	4112509996,
	4226922342,
	4269808519,
	4285694709,
	4291553771,
	4293711215,
	4294505167,
	4294797283,
	4294904752,
	4294944288,
	4294958833,
	4294964183,
	4294966151,
	4294966875,
}

const pTable4194304ZOff = 12

var pTable4194304 = [...]uint32{
	327810,
	131303775,
	1190527418,
	2678982827,
	3610341491,
	4029187631,
	4195212982,
	4257996890,
	4281329467,
	4289945175,
	4293119076,
	4294287280,
	4294717119,
	4294875257,
	4294933436,
	4294954838,
	4294962712,
	4294965608,
}

const pTable16777216ZOff = 13

var pTable16777216 = [...]uint32{
	3752,
	25355842,
	650129895,
	2144438260,
	3326521888,
	3909637329,
	4148984156,
	4240675262,
	4274914027,
	4287579196,
	4292247888,
	4293966687,
	4294599162,
	4294831868,
	4294917474,
	4294948970,
	4294960554,
}

const pTable67108864ZOff = 14

var pTable67108864 = [...]uint32{
	5,
	2254867,
	266915997,
	1545546076,
	2948942510,
	3740137088,
	4081882345,
	4215314238,
	4265491052,
	4284099979,
	4290966248,
	4293494940,
	4294425602,
	4294768004,
	4294893984,
	4294940319,
}

const pTable268435456ZOff = 16

var pTable268435456 = [...]uint32{
	64065,
	72019192,
	954521336,
	2469849553,
	3503985876,
	3985109308,
	4178270546,
	4251662734,
	4278985456,
	4289080901,
	4292800930,
	4294170185,
	4294674057,
	4294859392,
}

const pTable1073741824ZOff = 17

var pTable1073741824 = [...]uint32{
	340,
	10477629,
	469680377,
	1902697748,
	3183327853,
	3846862347,
	4124351229,
	4231395872,
	4271470080,
	4286308390,
	4291779745,
	4293794466,
	4294535704,
}

const pTable4294967296ZOff = 19

var pTable4294967296 = [...]uint32{
	614229,
	165424424,
	1296122030,
	2764057036,
	3652102740,
	4046271760,
	4201746919,
	4260436266,
	4282231273,
	4290277890,
	4293241179,
}

const pTable17179869184ZOff = 20

var pTable17179869184 = [...]uint32{
	9452,
	35620826,
	736730107,
	2245392988,
	3383300774,
	3934050924,
	4158500016,
	4244249356,
	4276240371,
	4288066990,
}

const pTable68719476736ZOff = 21

var pTable68719476736 = [...]uint32{
	20,
	3718382,
	320841001,
	1653799469,
	3023293602,
	3774567931,
	4095660799,
	4220548086,
	4267432512,
}

const pTable274877906944ZOff = 23

var pTable274877906944 = [...]uint32{
	133746,
	94417351,
	1054488182,
	2562060611,
	3551536989,
	4004938279,
	4185882787,
}

const pTable1099511627776ZOff = 24

var pTable1099511627776 = [...]uint32{
	1003,
	15605845,
	543848600,
	2008102075,
	3247160282,
	3874972893,
}

const pTable4398046511104ZOff = 25

var pTable4398046511104 = [...]uint32{
	1,
	1104161,
	205240732,
	1403250873,
	2845739169,
}
