# D&D Weapon Generator
This is a simple tool to generate random D&D weapons using a basic Go script.
"Random" weapons basically take two components: a "frame" with a basic description, 
hit chance, damage roll, and range, and adds a set of "perks" to it. These perks are
entirely homebrew and are, frankly, just random things I've made up. Once perks have
been added, a search is run to replace any randomization keys with random values.

For example, suppose we have a weapon with `"hit": "1d20+<1-6>"`, and in the `randoms` section
we define `"<1-6>" : ["1","2","3","4","5","6"]`. The resulting weapon may have `"hit": "1d20+1"`
or `"hit": "1d20+2"` and so on and so forth. These random selections are all unique (i.e., if
we have `"damage": "<1-6>d4+<1-6>"`, the program does not necessarily set each instance of `<1-6>`
to the same value).

I mostly made this program so that I only have to make some new JSONs if I want to give
my players some new tools for a new session, rather than randomly doing these things by hand. 
For the sake of balancing, I suggest keeping a close eye on how certain values (e.g., damage)
are randomized.

### Prerequisites
Have `go` installed. `sudo apt install golang-go` should work on most Ubuntu systems. For Windows,
installing from the webpage should be sufficient.

### Flags:
- --fin - File to use for perk, weapon, and random selection.
- --ffmt - File to use for formatting.
- --ngen - Number of weapons to generate.
- --nperk - Number of perks to add to each generated weapon.

### Example:
From the root of the repository, run:
```
go run src/main.go --fin examples/basic.json --ngen 5 --nperk 2 --ffmt formats/discord.txt
```
This will generate five weapons, each with two perks, and print them using the Discord format.


### Things to do:
- Allow different files for perks and weapons
- Give file access flags better defaults

Input files should have the following setup:
```
{
  "weapons": [
    {
      "name": "Name of your weapon",
      "hit": "Hit roll for the weapon, e.g. 1d20+4",
      "damage": "Damage roll for the weapon, e.g. 1d10",
      "range": "Range of the weapon, e.g. 20/80, 40ft., or melee.",
      "description": "A description of the weapon."
    }
  ],
  "perks": [
    "Any interesting tools a weapon could use."
  ],
  "randoms": {
    "key to look up" : [
        "random item to select",
        "other random item",
        "another random item"
    ]
  }
}
```

### Discord Format (formats/discord.txt)

```
# Montante:
    **Hit roll:** 4d8-1d6
    **Damage roll:** 1d12
    **Range:** Melee
    **Perks:**
        Hits half the target's movement on their next turn.
        Hits add 1d6 to damage rolls, to a max of 5d6. Returns to 0d6 on miss.
    **Description:** An overdesigned greatsword.
```

### Plaintext Format (formats/plaintext.txt)

```
Rapier:
    Hit roll:    2d10+8
    Damage roll: 1d4+1d6
    Range:       Melee
    Perks:
        Advantage hit rolls also do advantage damage.
        Hits do not count as an action. Once per turn.
    Description: An elegant rapier.
```