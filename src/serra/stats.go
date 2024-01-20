package serra

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Aliases:       []string{"stats"},
	Use:           "stats <prefix> <n>",
	Short:         "Shows statistics of the collection",
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		client := storageConnect()
		coll := &Collection{client.Database("serra").Collection("cards")}
		totalcoll := &Collection{client.Database("serra").Collection("total")}
		l := Logger()
		defer storageDisconnect(client)

		// Value and Card Numbers
		stats, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "value", Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								getCurrencyField(false),
								"$serra_count",
							}},
						}},
					}},
					{Key: "value_foil", Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								getCurrencyField(true),
								"$serra_count_foil",
							}},
						}},
					}},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								1.0,
								"$serra_count",
							}},
						}},
					}},
					{Key: "count_foil", Value: bson.D{
						{Key: "$sum", Value: "$serra_count_foil"},
					}},
					{Key: "rarity", Value: bson.D{
						{Key: "$sum", Value: "$rarity"},
					}},
					{Key: "unique", Value: bson.D{
						{Key: "$sum", Value: 1},
					}},
				}},
			},
			bson.D{
				{Key: "$addFields", Value: bson.D{
					{Key: "count_all", Value: bson.D{
						{Key: "$sum", Value: bson.A{
							"$count",
							"$count_foil",
						}},
					}},
				}},
			},
		})
		fmt.Printf("%sCards %s\n", Green, Reset)
		fmt.Printf("Total: %s%.0f%s\n", Yellow, stats[0]["count_all"], Reset)
		fmt.Printf("Unique: %s%d%s\n", Purple, stats[0]["unique"], Reset)
		fmt.Printf("Normal: %s%.0f%s\n", Purple, stats[0]["count"], Reset)
		fmt.Printf("Foil: %s%d%s\n", Purple, stats[0]["count_foil"], Reset)

		reserved, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "reserved", Value: true},
				}},
			},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: 1},
					}},
				}},
			},
		})

		var count_reserved int32
		if len(reserved) > 0 {
			count_reserved = reserved[0]["count"].(int32)
		}
		fmt.Printf("Reserved List: %s%d%s\n", Yellow, count_reserved, Reset)

		// Rarities
		rar, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$rarity"},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								1.0,
								"$serra_count",
							}},
						}},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "_id", Value: 1},
				}},
			},
		})
		ri := convertRarities(rar)
		fmt.Printf("\n%sRarity%s\n", Green, Reset)
		fmt.Printf("Mythics: %s%.0f%s\n", Pink, ri.Mythics, Reset)
		fmt.Printf("Rares: %s%.0f%s\n", Pink, ri.Rares, Reset)
		fmt.Printf("Uncommons: %s%.0f%s\n", Yellow, ri.Uncommons, Reset)
		fmt.Printf("Commons: %s%.0f%s\n", Purple, ri.Commons, Reset)

		// Colors
		sets, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "coloridentity", Value: bson.D{
						{Key: "$size", Value: 1},
					}},
				}},
			},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$coloridentity"},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$multiply", Value: bson.A{
								1.0,
								"$serra_count",
							}},
						}},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "count", Value: -1},
				}},
			},
		})

		fmt.Printf("\n%sColors%s\n", Green, Reset)
		for _, set := range sets {
			x, _ := set["_id"].(primitive.A)
			s := []interface{}(x)
			fmt.Printf("%s: %s%.0f%s\n", convertManaSymbols(s), Purple, set["count"], Reset)
		}
		// Artists
		artists, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$artist"},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: 1},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "count", Value: -1},
				}},
			},
			bson.D{
				{Key: "$limit", Value: 10},
			},
		})
		fmt.Printf("\n%sTop Artists%s\n", Green, Reset)
		for _, artist := range artists {
			fmt.Printf("%s: %s%d%s\n", artist["_id"].(string), Purple, artist["count"], Reset)
		}

		// Mana Curve of Collection
		cmc, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$cmc"},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: 1},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "_id", Value: 1},
				}},
			},
		})
		fmt.Printf("\n%sMana Curve%s\n", Green, Reset)
		for _, mc := range cmc {
			fmt.Printf("%.0f: %s%d%s\n", mc["_id"], Purple, mc["count"], Reset)
		}

		// Show cards added per month
		fmt.Printf("\n%sCards added over time%s\n", Green, Reset)
		type Caot struct {
			Id struct {
				Year  int32 `mapstructure:"year"`
				Month int32 `mapstructure:"month"`
			} `mapstructure:"_id"`
			Count int32 `mapstructure:"count"`
		}
		caot, _ := coll.storageAggregate(mongo.Pipeline{
			bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "month", Value: bson.D{
						{Key: "$month", Value: "$serra_created"},
					}},
					{Key: "year", Value: bson.D{
						{Key: "$year", Value: "$serra_created"},
					}},
				}},
			},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: bson.D{
						{Key: "month", Value: "$month"},
						{Key: "year", Value: "$year"},
					}},
					{Key: "count", Value: bson.D{
						{Key: "$sum", Value: 1},
					}},
				}},
			},
			bson.D{
				{Key: "$sort", Value: bson.D{
					{Key: "_id.year", Value: 1},
					{Key: "_id.month", Value: 1},
				}},
			},
		})
		for _, mo := range caot {
			moo := new(Caot)
			mapstructure.Decode(mo, moo)
			fmt.Printf("%d-%02d: %s%d%s\n", moo.Id.Year, moo.Id.Month, Purple, moo.Count, Reset)
		}

		// Total Value
		fmt.Printf("\n%sTotal Value%s\n", Green, Reset)
		normalValue, err := getFloat64(stats[0]["value"])
		if err != nil {
			l.Error(err)
			normalValue = 0
		}
		foilValue, err := getFloat64(stats[0]["value_foil"])
		if err != nil {
			l.Error(err)
			foilValue = 0
		}
		count_all, err := getFloat64(stats[0]["count_all"])
		if err != nil {
			l.Error(err)
			foilValue = 0
		}
		totalValue := normalValue + foilValue
		averageValue := 0.0
		if count_all > 0 {
			averageValue = totalValue/count_all 
		}
		fmt.Printf("Total: %s%.2f%s%s\n", Pink, totalValue, getCurrency(), Reset)
		fmt.Printf("Normal: %s%.2f%s%s\n", Pink, normalValue, getCurrency(), Reset)
		fmt.Printf("Foils: %s%.2f%s%s\n", Pink, foilValue, getCurrency(), Reset)
		fmt.Printf("Average Card: %s%.2f%s%s\n", Pink, averageValue, getCurrency(), Reset)
		total, err := totalcoll.storageFindTotal()
		if err == nil {
			fmt.Printf("History: \n")
			showPriceHistory(total.Value, "* ", true)
		}
		return nil
	},
}
