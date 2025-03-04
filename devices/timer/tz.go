package timer

import (
	"sort"
)

var tzMap = map[string]string{
	"PST":   "-0800", // Pacific Standard Time
	"PDT":   "-0700", // Pacific Daylight Time
	"EST":   "-0500", // Eastern Standard Time
	"EDT":   "-0400", // Eastern Daylight Time
	"UTC":   "+0000", // UTC
	"IDLW":  "-1200", // International Date Line West
	"SST":   "-1100", // Samoa Standard Time
	"HST":   "-1000", // Hawaii Standard Time
	"TAHT":  "-1000", // Tahiti Time
	"MART":  "-0930", // Marquesas Time
	"AKST":  "-0900", // Alaska Standard Time
	"GAMT":  "-0900", // Gambier Time
	"PIT":   "-0800", // Pitcairn Time
	"MST":   "-0700", // Mountain Standard Time
	"MeST":  "-0700", // Mexico Standard Time
	"MDT":   "-0600", // Mountain Daylight Time
	"CST":   "-0600", // Central Standard Time (US)
	"CAST":  "-0600", // Central America Standard Time
	"CDT":   "-0500", // Central Daylight Time
	"ACT":   "-0500", // Acre Time
	"COT":   "-0500", // Colombia Time
	"PET":   "-0500", // Peru Time
	"AST":   "-0400", // Atlantic Standard Time
	"CLT":   "-0400", // Chile Standard Time
	"AMT":   "-0400", // Amazon Time
	"BOT":   "-0400", // Bolivia Time
	"NST":   "-0330", // Newfoundland Standard Time
	"NDT":   "-0230", // Newfoundland Daylight Time
	"ADT":   "-0300", // Atlantic Daylight Time
	"BRT":   "-0300", // Brasília Time
	"ART":   "-0300", // Argentina Time
	"CLST":  "-0300", // Chile Summer Time
	"FKT":   "-0300", // Falkland Islands Time
	"SRT":   "-0300", // Suriname Time
	"UYT":   "-0300", // Uruguay Time
	"BRST":  "-0200", // Brasília Summer Time
	"UYST":  "-0200", // Uruguay Summer Time
	"GST":   "-0200", // South Georgia Time (also +0400 for Gulf)
	"AZOT":  "-0100", // Azores Standard Time
	"CVT":   "-0100", // Cape Verde Time
	"WET":   "+0000", // Western European Time
	"AZOST": "+0000", // Azores Summer Time
	"GMT":   "+0000", // Greenwich Mean Time
	"WEST":  "+0100", // Western European Summer Time
	"CET":   "+0100", // Central European Time
	"WAT":   "+0100", // West Africa Time
	"CEST":  "+0200", // Central European Summer Time
	"EET":   "+0200", // Eastern European Time
	"CAT":   "+0200", // Central Africa Time
	"SAST":  "+0200", // South Africa Standard Time
	"IST":   "+0200", // Israel Standard Time
	"EEST":  "+0300", // Eastern European Summer Time
	"IDT":   "+0300", // Israel Daylight Time
	"MSK":   "+0300", // Moscow Standard Time
	"EAT":   "+0300", // East Africa Time
	"IRST":  "+0330", // Iran Standard Time
	"SAMT":  "+0400", // Samara Time
	"AFT":   "+0430", // Afghanistan Time
	"PKT":   "+0500", // Pakistan Standard Time
	"UZT":   "+0500", // Uzbekistan Time
	"YEKT":  "+0500", // Yekaterinburg Time
	"NPT":   "+0545", // Nepal Time
	"BST":   "+0600", // Bangladesh Standard Time
	"OMST":  "+0600", // Omsk Standard Time
	"CCT":   "+0630", // Cocos Islands Time
	"MMT":   "+0630", // Myanmar Time
	"ICT":   "+0700", // Indochina Time
	"KRAT":  "+0700", // Krasnoyarsk Time
	"WIB":   "+0700", // Western Indonesian Time
	"HKT":   "+0800", // Hong Kong Time
	"SGT":   "+0800", // Singapore Time
	"AWST":  "+0800", // Australian Western Standard Time
	"WITA":  "+0800", // Central Indonesian Time
	"IRKT":  "+0800", // Irkutsk Time
	"CWST":  "+0845", // Central Western Standard Time
	"JST":   "+0900", // Japan Standard Time
	"KST":   "+0900", // Korea Standard Time
	"WIT":   "+0900", // Eastern Indonesian Time
	"YAKT":  "+0900", // Yakutsk Time
	"AEST":  "+1000", // Australian Eastern Standard Time
	"PGT":   "+1000", // Papua New Guinea Time
	"VLAT":  "+1000", // Vladivostok Time
	"ACDT":  "+1030", // Australian Central Daylight Time
	"LHST":  "+1030", // Lord Howe Standard Time
	"AEDT":  "+1100", // Australian Eastern Daylight Time
	"SBT":   "+1100", // Solomon Islands Time
	"NCT":   "+1100", // New Caledonia Time
	"VUT":   "+1100", // Vanuatu Time
	"MAGT":  "+1100", // Magadan Time
	"NFT":   "+1130", // Norfolk Island Time
	"NZST":  "+1200", // New Zealand Standard Time
	"FJT":   "+1200", // Fiji Time
	"KALT":  "+1200", // Kamchatka Time
	"CHAST": "+1245", // Chatham Standard Time
	"NZDT":  "+1300", // New Zealand Daylight Time
	"TOT":   "+1300", // Tonga Time
	"PHOT":  "+1300", // Phoenix Island Time
	"LINT":  "+1400", // Line Islands Time
}

func timeZones() []string {
	// Collect all keys from tzMap
	keys := make([]string, 0, len(tzMap))
	for key := range tzMap {
		keys = append(keys, key)
	}

	// Sort the keys alphabetically
	sort.Strings(keys)
	return keys
}
