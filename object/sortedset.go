package object

type ZSet interface {

	// ZADD adds one or more members in the sorted set
	ZADD(...ZSetMember) error

	// ZCARD gets the number of members in the sorted set
	ZCARD() int

	// ZCOUNT count the members with scores within the given values in the sorted set
	ZCOUNT(min, max float64) int

	// ZDIFF subtracts the sorted set and other multiple sorted sets
	ZDIFF(...ZSet) []ZSetMember

	// ZDIFFSTORE subtracts the sorted set and other multiple sorted sets and stores it
	// ZDIFFSTORE() ZSet
	
	ZINCRBY(key string) error
	
	// ZINTER intersect the sorted set and other multiple sorted sets
	ZINTER()

	// ZINTERSTORE()
	// ZLEXOUT()

	ZRANGEBYSCORE(min, max float64) []ZSetMember

	ZRANGEBYLEX(min, max string) []ZSetMember

	ZRANK(key string) int

	ZSCORE(key string) float64

	ZPOPMAX() ZSetMember
	ZPOPMIN() ZSetMember

}

type ZSetMember struct {
	Key string
	Value float64
}