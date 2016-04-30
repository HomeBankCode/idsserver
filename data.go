package main

/*
DataGroup is a collection of paths to
CLAN files and blocks (audio clips) which
are going to be served to users.
*/
type DataGroup struct {
	ClanFile string
	Blocks   []string
}

/*
DataManifest is a map of integer ID's to
DataGroups. DataGroups give info for path
lookups to the relevant files
*/
type DataManifest map[uint]DataGroup

/*
DataActive is a map of ID's to booleans.
The bool values represent whether a
CLAN file is actively being worked on
or not.
*/
type DataActive map[uint]bool
