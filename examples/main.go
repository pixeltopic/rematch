package main

import (
	"fmt"

	"github.com/pixeltopic/rematch"
)

const example = `
Apples ducks straw, quail a ostriches donkey, hay hook cucumbers. 
Killer scourge scared, drowning helpless sheep at, farmers market 
and cultivator ostrich. Combine Harvester swather, baler as haybine 
parsley, melon in hay rake. Forage Harvester rakes peacocks, 
squeal garden woof. Goose hammers cattle rats in crows. House hen 
chinchillas in barn livestock cat hogs chicks trucks. Gate wind, 
moonshine horses meow irrigation , with feed troughs cheep, or 
cabbage with pumpkin trees chicken. Fee.

In a woof, a farmers market. Shovels at rakes plows. Gourds 
utters at welding equipment a oink oink haybine. Forage Harvester 
rakes peacocks, squeal garden woof. Post pounder calf, hay or duck 
is, tool shed horse.`

func main() {

	rawExprs := []string{
		"ostriches+Apples+horse",           // true
		"apples+ostriches",                 // false
		"apples|ostriches",                 // true
		"Apples|ostriches+apples",          // false
		"Apples|(ostriches+apples)",        // true
		"((Apples)|((ostriches+apples)))",  // true (equivalent to previous)
		"scared*sheep",                     // true
		"scared?sheep",                     // false
		"livestock_cat_hogs_chicks_trucks", // true
		"!jolly_cow",                       // true
		"wind*moonshine",                   // true
	}

	txt := rematch.NewText(example)

	for _, e := range rawExprs {
		expr := rematch.NewExpr(e)
		res, _ := rematch.FindAll(expr, txt)
		fmt.Printf("expr '%s' result: %v %v\n", expr.Raw(), res.Match, res.Strings)
	}
}
