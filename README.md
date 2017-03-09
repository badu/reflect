# reflector

Utility for reflecting structs

# why

Reflection is hard.

# what is doing so far

- [x] inspects a struct, traversing it (no pointers)
- [x] collects fields inside all structs
- [x] establishes relations between models
- [x] collects tags for each field and keeps them in key-value-option struct
- [x] checks for time.Time implementation on fields
- [x] works with embedded fields
- [x] checks for structs that self references themselves
- [x] given a white list of methods, will create a map of string-function inside the model
- [ ] establishes relation's kind for each field-model pair (to be described)
- [ ] check for Scanner, Valuer, Marshal and Unmarshal implementation on structs
