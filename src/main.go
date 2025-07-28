package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

// These constants handle flag titles via CLI.
const INPUT_FILE_FLAG = "fin"   // Input file for data
const FORMAT_FILE_FLAG = "ffmt" // Input file for format
const NGEN_FLAG = "ngen"        // Number of generated weapons
const NPERK_FLAG = "nperk"      // Number of selected perks per weapon

// These constants handle what strings are replaced with what data in a
// given input format file.
const REPLACE_NAME_STR = "{REPLACE_NAME_STR}"
const REPLACE_HIT_STR = "{REPLACE_HIT_STR}"
const REPLACE_DAMAGE_STR = "{REPLACE_DAMAGE_STR}"
const REPLACE_RANGE_STR = "{REPLACE_RANGE_STR}"
const REPLACE_PERK_STR = "{REPLACE_PERK_STR}"
const REPLACE_DESC_STR = "{REPLACE_DESC_STR}"

// This struct handles individual flags.
type flag_t struct {
	fname    string  // Name of the flag.
	fdefault any     // Default value. Proper assignment allows functions to determine type.
	fhelp    string  // Simple help string for CLI use
	fstring  *string // If the type of fdefault is a string, the value is placed here.
	fint     *int    // If the type of fdefault is an int, the value is placed here.
}

// All flags in a struct, for simpler passing around between functions.
type all_flags_t struct {
	fin   string // INPUT_FILE_FLAG
	ffmt  string // FORMAT_FILE_FLAG
	ngen  int    // NGEN_FLAG
	nperk int    // NPERK_FLAG
}

type weapon_base_t struct {
	W_name   string `json:"name"`
	W_hit    string `json:"hit"`
	W_damage string `json:"damage"`
	W_range  string `json:"range"`
	W_desc   string `json:"description"`
}

type weapon_generated_t struct {
	weapon weapon_base_t
	perks  []string
}

type input_file_t struct {
	I_weapons []weapon_base_t     `json:"weapons"`
	I_perks   []string            `json:"perks"`
	I_rands   map[string][]string `json:"randoms"`
}

func parse_flags(fmap *map[string]*flag_t) {
	for fkey, fval := range *fmap {
		if fdefault, ok := fval.fdefault.(string); ok {
			(*fmap)[fkey].fstring = flag.String(fval.fname, fdefault, fval.fhelp)
		} else if fdefault, ok := fval.fdefault.(int); ok {
			(*fmap)[fkey].fint = flag.Int(fval.fname, fdefault, fval.fhelp)
		} else {
			fmt.Printf("Unsupported flag type! Discarding info...\n")
		}
	}
	flag.Parse()
}

func add_flag(fmap *map[string]*flag_t, fname string, fdefault any, fhelp string) {
	(*fmap)[fname] = &flag_t{fname: fname, fdefault: fdefault, fhelp: fhelp}
}

func build_and_read_flags() all_flags_t {
	var flags_out all_flags_t

	fmap := make(map[string]*flag_t)
	add_flag(&fmap, INPUT_FILE_FLAG, "", "Input file for generated stats, names, and perks.")
	add_flag(&fmap, FORMAT_FILE_FLAG, "", "Format file for generated weapons.")
	add_flag(&fmap, NGEN_FLAG, 1, "Number of weapons to generate")
	add_flag(&fmap, NPERK_FLAG, 1, "Number of perks to allocate to each weapon.")
	parse_flags(&fmap)

	flags_out.fin = *fmap[INPUT_FILE_FLAG].fstring
	flags_out.ffmt = *fmap[FORMAT_FILE_FLAG].fstring
	flags_out.ngen = *fmap[NGEN_FLAG].fint
	flags_out.nperk = *fmap[NPERK_FLAG].fint

	return flags_out
}

func read_file(flags_out all_flags_t) input_file_t {
	var input_file input_file_t
	fin_file, _ := os.ReadFile(flags_out.fin)
	json.Unmarshal(fin_file, &input_file)
	return input_file
}

func insert_randoms_string(str string, rands map[string][]string) string {
	// Check for each random key in the string and replace as directed.
	for key, value := range rands {
		// Continue randomization while str contains key
		for strings.Contains(str, key) {
			// Pick a random string from the given list and replace one instance
			//	of the key with it. This ensures that multiple instances of one
			//	operation can be replaced by multiple different options.
			rand := string(value[rand.IntN(len(rands[key]))])
			str = strings.Replace(str, key, rand, 1)
		}
	}
	return str
}

func insert_randoms_weapon(weapon weapon_generated_t, rands map[string][]string) weapon_generated_t {
	// These objects are individual strings, so we only need to handle randoms once each.
	weapon.weapon.W_name = insert_randoms_string(weapon.weapon.W_name, rands) // Reasonably should never be randomized.
	weapon.weapon.W_damage = insert_randoms_string(weapon.weapon.W_damage, rands)
	weapon.weapon.W_hit = insert_randoms_string(weapon.weapon.W_hit, rands)
	weapon.weapon.W_range = insert_randoms_string(weapon.weapon.W_range, rands)
	weapon.weapon.W_desc = insert_randoms_string(weapon.weapon.W_desc, rands) // Reasonably should never be randomized.

	// "Perks" is a list, so we need to loop through each element and update them individually.
	for index, value := range weapon.perks {
		weapon.perks[index] = insert_randoms_string(value, rands)
	}

	// The returned weapon should be 100% generated at this point.
	return weapon
}

func build_random(nperk int, input_file input_file_t) weapon_generated_t {
	// Randomly select a weapon. No bias / exclusion necessary.
	weapon := input_file.I_weapons[rand.IntN(len(input_file.I_weapons))]

	// Get and check the number of perks. We can't have more selected perks than available perks.
	num_possible_perks := len(input_file.I_perks)
	if num_possible_perks < nperk {
		fmt.Printf("%d perks requested, but only %d perks exist. Providing %d perks.\n", nperk, num_possible_perks, num_possible_perks)
		nperk = num_possible_perks
	}

	// This block handles selecting a set of perks. We create a set (map of int pointers, where
	// 	all int pointers are nil) to keep track of which perks we already have, and end up with a list
	//	of all selected perks. The list is not ordered (i.e., it is equally likely to have perks X,Y,Z
	//	perks Y,X,Z).
	perk_set := make(map[int]*int)
	var perks []string
	for n := 0; n < nperk; {
		// Generate a random number. If that perk already exists in the set, generate a random number
		//	again. This may take a while in situations where the number of perks is very large, and the
		//	number of selected perks is very close to that number, but those situations are unrealistic.
		randn := rand.IntN(num_possible_perks)
		if _, ok := perk_set[randn]; ok {
			continue
		}
		n++

		// Append the selected perk to the list for usage, add the perk number to the set.
		perks = append(perks, input_file.I_perks[randn])
		perk_set[randn] = nil
	}

	// Final call before returning the generated weapon handles generating random values where needed.
	return insert_randoms_weapon(weapon_generated_t{weapon: weapon, perks: perks}, input_file.I_rands)
}

// This function is just a simple list to make repeated calls to build_random() for weapon generation.
func generate_weapons(flags_out all_flags_t, input_file input_file_t) []weapon_generated_t {
	var generated_weapons []weapon_generated_t
	for range flags_out.ngen {
		generated_weapons = append(generated_weapons, build_random(flags_out.nperk, input_file))
	}
	return generated_weapons
}

func generate_individual_output(ffmt string, weapon weapon_generated_t) string {
	// Since these groups only have one "section" we can just run an old fashioned replace.
	replaced := strings.Replace(ffmt, REPLACE_NAME_STR, weapon.weapon.W_name, -1)
	replaced = strings.Replace(replaced, REPLACE_HIT_STR, weapon.weapon.W_hit, -1)
	replaced = strings.Replace(replaced, REPLACE_DAMAGE_STR, weapon.weapon.W_damage, -1)
	replaced = strings.Replace(replaced, REPLACE_RANGE_STR, weapon.weapon.W_range, -1)
	replaced = strings.Replace(replaced, REPLACE_DESC_STR, weapon.weapon.W_desc, -1)

	// Since there can be a lot of perks, we need to copy any line that wants perks
	//	and paste it for however many perks we actually have. This, of course, therefore assumes
	//	that each perk gets its own line, but I think that's just common sense.
	replaced_split := strings.Split(replaced, "\n")
	for index, line := range replaced_split {
		if !strings.Contains(line, REPLACE_PERK_STR) {
			continue
		}

		var updated_line []string
		for _, perk := range weapon.perks {
			updated_line = append(updated_line, strings.Replace(line, REPLACE_PERK_STR, perk, -1))
		}

		replaced_split[index] = strings.Join(updated_line, "\n")
	}
	replaced = strings.Join(replaced_split, "\n")
	return replaced
}

func generate_output(flags_out all_flags_t, weapons []weapon_generated_t) {
	ffmt_raw, _ := os.ReadFile(flags_out.ffmt)
	ffmt := string(ffmt_raw)
	for _, weapon := range weapons {
		fmt.Printf("%s\n", generate_individual_output(ffmt, weapon))
	}
}

func main() {
	// Parse input flags and generate our command struct using them.
	flags_out := build_and_read_flags()

	// Read the input json data file into a simple data struct.
	input_file := read_file(flags_out)

	// Actually perform generation.
	generated_weapons := generate_weapons(flags_out, input_file)

	// Format output and send to stdout.
	generate_output(flags_out, generated_weapons)
}
