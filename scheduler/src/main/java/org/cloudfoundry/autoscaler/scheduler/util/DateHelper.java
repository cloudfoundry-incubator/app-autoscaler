package org.cloudfoundry.autoscaler.scheduler.util;

import java.text.SimpleDateFormat;
import java.time.Instant;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.time.ZonedDateTime;
import java.util.Date;
import java.util.TimeZone;

public class DateHelper {

	public static final String DATE_TIME_FORMAT = "yyyy-MM-dd'T'HH:mm";
	public static final String DATE_FORMAT = "yyyy-MM-dd";
	public static final String TIME_FORMAT = "HH:mm";

	public static final int DAY_OF_WEEK_MINIMUM = 1;
	public static final int DAY_OF_WEEK_MAXIMUM = 7;
	public static final int DAY_OF_MONTH_MINIMUM = 1;
	public static final int DAY_OF_MONTH_MAXIMUM = 31;

	static final String[] supportedTimezones = new String[] {
	           "Etc/GMT+12",
	           "Etc/GMT+11",
	           "Pacific/Midway",
	           "Pacific/Niue",
	           "Pacific/Pago_Pago",
	           "Pacific/Samoa",
	           "US/Samoa",
	           "Etc/GMT+10",
	           "HST",
	           "Pacific/Honolulu",
	           "Pacific/Johnston",
	           "Pacific/Rarotonga",
	           "Pacific/Tahiti",
	           "US/Hawaii",
	           "Pacific/Marquesas",
	           "America/Adak",
	           "America/Atka",
	           "Etc/GMT+9",
	           "Pacific/Gambier",
	           "US/Aleutian",
	           "America/Anchorage",
	           "America/Juneau",
	           "America/Metlakatla",
	           "America/Nome",
	           "America/Sitka",
	           "America/Yakutat",
	           "Etc/GMT+8",
	           "Pacific/Pitcairn",
	           "US/Alaska",
	           "America/Creston",
	           "America/Dawson",
	           "America/Dawson_Creek",
	           "America/Ensenada",
	           "America/Hermosillo",
	           "America/Los_Angeles",
	           "America/Phoenix",
	           "America/Santa_Isabel",
	           "America/Tijuana",
	           "America/Vancouver",
	           "America/Whitehorse",
	           "Canada/Pacific",
	           "Canada/Yukon",
	           "Etc/GMT+7",
	           "MST",
	           "Mexico/BajaNorte",
	           "PST8PDT",
	           "US/Arizona",
	           "US/Pacific",
	           "US/Pacific-New",
	           "America/Belize",
	           "America/Boise",
	           "America/Cambridge_Bay",
	           "America/Chihuahua",
	           "America/Costa_Rica",
	           "America/Denver",
	           "America/Edmonton",
	           "America/El_Salvador",
	           "America/Guatemala",
	           "America/Inuvik",
	           "America/Managua",
	           "America/Mazatlan",
	           "America/Ojinaga",
	           "America/Regina",
	           "America/Shiprock",
	           "America/Swift_Current",
	           "America/Tegucigalpa",
	           "America/Yellowknife",
	           "Canada/East-Saskatchewan",
	           "Canada/Mountain",
	           "Canada/Saskatchewan",
	           "Etc/GMT+6",
	           "MST7MDT",
	           "Mexico/BajaSur",
	           "Navajo",
	           "Pacific/Galapagos",
	           "US/Mountain",
	           "America/Atikokan",
	           "America/Bahia_Banderas",
	           "America/Bogota",
	           "America/Cancun",
	           "America/Cayman",
	           "America/Chicago",
	           "America/Coral_Harbour",
	           "America/Eirunepe",
	           "America/Guayaquil",
	           "America/Indiana/Knox",
	           "America/Indiana/Tell_City",
	           "America/Jamaica",
	           "America/Knox_IN",
	           "America/Lima",
	           "America/Matamoros",
	           "America/Menominee",
	           "America/Merida",
	           "America/Mexico_City",
	           "America/Monterrey",
	           "America/North_Dakota/Beulah",
	           "America/North_Dakota/Center",
	           "America/North_Dakota/New_Salem",
	           "America/Panama",
	           "America/Porto_Acre",
	           "America/Rainy_River",
	           "America/Rankin_Inlet",
	           "America/Resolute",
	           "America/Rio_Branco",
	           "America/Winnipeg",
	           "Brazil/Acre",
	           "CST6CDT",
	           "Canada/Central",
	           "Chile/EasterIsland",
	           "EST",
	           "Etc/GMT+5",
	           "Jamaica",
	           "Mexico/General",
	           "Pacific/Easter",
	           "US/Central",
	           "US/Indiana-Starke",
	           "America/Caracas",
	           "America/Anguilla",
	           "America/Antigua",
	           "America/Aruba",
	           "America/Asuncion",
	           "America/Barbados",
	           "America/Blanc-Sablon",
	           "America/Boa_Vista",
	           "America/Campo_Grande",
	           "America/Cuiaba",
	           "America/Curacao",
	           "America/Detroit",
	           "America/Dominica",
	           "America/Fort_Wayne",
	           "America/Grand_Turk",
	           "America/Grenada",
	           "America/Guadeloupe",
	           "America/Guyana",
	           "America/Havana",
	           "America/Indiana/Indianapolis",
	           "America/Indiana/Marengo",
	           "America/Indiana/Petersburg",
	           "America/Indiana/Vevay",
	           "America/Indiana/Vincennes",
	           "America/Indiana/Winamac",
	           "America/Indianapolis",
	           "America/Iqaluit ",
	           "America/Kentucky/Louisville ",
	           "America/Kentucky/Monticello",
	           "America/Kralendijk",
	           "America/La_Paz",
	           "America/Louisville ",
	           "America/Lower_Princes",
	           "America/Manaus",
	           "America/Marigot",
	           "America/Martinique",
	           "America/Montreal",
	           "America/Montserrat",
	           "America/Nassau",
	           "America/New_York",
	           "America/Nipigon",
	           "America/Pangnirtung ",
	           "America/Port-au-Prince ",
	           "America/Port_of_Spain",
	           "America/Porto_Velho",
	           "America/Puerto_Rico ",
	           "America/Santo_Domingo ",
	           "America/St_Barthelemy",
	           "America/St_Kitts",
	           "America/St_Lucia",
	           "America/St_Thomas",
	           "America/St_Vincent",
	           "America/Thunder_Bay",
	           "America/Toronto",
	           "America/Tortola",
	           "America/Virgin",
	           "Brazil/West",
	           "Canada/Eastern",
	           "Cuba",
	           "EST5EDT",
	           "Etc/GMT+4",
	           "US/East-Indiana",
	           "US/Eastern",
	           "US/Michigan",
	           "America/Araguaina ",
	           "America/Argentina/Buenos_Aires ",
	           "America/Argentina/Catamarca ",
	           "America/Argentina/ComodRivadavia ",
	           "America/Argentina/Cordoba ",
	           "America/Argentina/Jujuy ",
	           "America/Argentina/La_Rioja ",
	           "America/Argentina/Mendoza ",
	           "America/Argentina/Rio_Gallegos ",
	           "America/Argentina/Salta ",
	           "America/Argentina/San_Juan ",
	           "America/Argentina/San_Luis ",
	           "America/Argentina/Tucuman ",
	           "America/Argentina/Ushuaia",
	           "America/Bahia",
	           "America/Belem",
	           "America/Buenos_Aires",
	           "America/Catamarca",
	           "America/Cayenne",
	           "America/Cordoba",
	           "America/Fortaleza",
	           "America/Glace_Bay",
	           "America/Goose_Bay",
	           "America/Halifax",
	           "America/Jujuy",
	           "America/Maceio",
	           "America/Mendoza",
	           "America/Moncton",
	           "America/Montevideo",
	           "America/Paramaribo",
	           "America/Recife",
	           "America/Rosario",
	           "America/Santarem",
	           "America/Santiago",
	           "America/Sao_Paulo",
	           "America/Thule",
	           "Antarctica/Palmer",
	           "Antarctica/Rothera",
	           "Atlantic/Bermuda",
	           "Atlantic/Stanley",
	           "Brazil/East",
	           "Canada/Atlantic",
	           "Chile/Continental",
	           "Etc/GMT+3",
	           "America/St_Johns",
	           "Canada/Newfoundland",
	           "America/Godthab",
	           "America/Miquelon",
	           "America/Noronha ",
	           "Atlantic/South_Georgia",
	           "Brazil/DeNoronha",
	           "Etc/GMT+2",
	           "Atlantic/Cape_Verde",
	           "Etc/GMT+1",
	           "Africa/Abidjan",
	           "Africa/Accra",
	           "Africa/Bamako",
	           "Africa/Banjul",
	           "Africa/Bissau",
	           "Africa/Conakry",
	           "Africa/Dakar",
	           "Africa/Freetown",
	           "Africa/Lome",
	           "Africa/Monrovia",
	           "Africa/Nouakchott",
	           "Africa/Ouagadougou",
	           "Africa/Sao_Tome",
	           "Africa/Timbuktu",
	           "America/Danmarkshavn",
	           "America/Scoresbysund",
	           "Atlantic/Azores",
	           "Atlantic/Reykjavik",
	           "Atlantic/St_Helena",
	           "Etc/GMT",
	           "Etc/GMT+0",
	           "Etc/GMT-0",
	           "Etc/GMT0",
	           "Etc/Greenwich",
	           "Etc/UCT",
	           "Etc/UTC",
	           "Etc/Universal",
	           "Etc/Zulu",
	           "GMT",
	           "GMT+0",
	           "GMT-0",
	           "GMT0",
	           "Greenwich",
	           "Iceland",
	           "UCT",
	           "UTC",
	           "Universal",
	           "Zulu",
	           "Africa/Algiers",
	           "Africa/Bangui",
	           "Africa/Brazzaville",
	           "Africa/Casablanca",
	           "Africa/Douala",
	           "Africa/El_Aaiun",
	           "Africa/Kinshasa",
	           "Africa/Lagos",
	           "Africa/Libreville",
	           "Africa/Luanda",
	           "Africa/Malabo",
	           "Africa/Ndjamena",
	           "Africa/Niamey",
	           "Africa/Porto-Novo",
	           "Africa/Tunis",
	           "Africa/Windhoek",
	           "Atlantic/Canary",
	           "Atlantic/Faeroe",
	           "Atlantic/Faroe",
	           "Atlantic/Madeira",
	           "Eire",
	           "Etc/GMT-1",
	           "Europe/Belfast",
	           "Europe/Dublin",
	           "Europe/Guernsey",
	           "Europe/Isle_of_Man",
	           "Europe/Jersey",
	           "Europe/Lisbon",
	           "Europe/London",
	           "GB",
	           "GB-Eire",
	           "Portugal",
	           "WET",
	           "Africa/Blantyre",
	           "Africa/Bujumbura",
	           "Africa/Cairo",
	           "Africa/Ceuta",
	           "Africa/Gaborone",
	           "Africa/Harare",
	           "Africa/Johannesburg",
	           "Africa/Kigali",
	           "Africa/Lubumbashi",
	           "Africa/Lusaka",
	           "Africa/Maputo",
	           "Africa/Maseru",
	           "Africa/Mbabane",
	           "Africa/Tripoli",
	           "Antarctica/Troll",
	           "Arctic/Longyearbyen",
	           "Atlantic/Jan_Mayen",
	           "CET",
	           "Egypt",
	           "Etc/GMT-2",
	           "Europe/Amsterdam",
	           "Europe/Andorra",
	           "Europe/Belgrade",
	           "Europe/Berlin",
	           "Europe/Bratislava",
	           "Europe/Brussels",
	           "Europe/Budapest",
	           "Europe/Busingen",
	           "Europe/Copenhagen",
	           "Europe/Gibraltar",
	           "Europe/Kaliningrad",
	           "Europe/Ljubljana",
	           "Europe/Luxembourg",
	           "Europe/Madrid",
	           "Europe/Malta",
	           "Europe/Monaco",
	           "Europe/Oslo",
	           "Europe/Paris",
	           "Europe/Podgorica",
	           "Europe/Prague",
	           "Europe/Rome",
	           "Europe/San_Marino",
	           "Europe/Sarajevo",
	           "Europe/Skopje",
	           "Europe/Stockholm",
	           "Europe/Tirane",
	           "Europe/Vaduz",
	           "Europe/Vatican",
	           "Europe/Vienna",
	           "Europe/Warsaw",
	           "Europe/Zagreb",
	           "Europe/Zurich",
	           "Libya",
	           "MET",
	           "Poland",
	           "Africa/Addis_Ababa",
	           "Africa/Asmara",
	           "Africa/Asmera",
	           "Africa/Dar_es_Salaam",
	           "Africa/Djibouti",
	           "Africa/Juba",
	           "Africa/Kampala",
	           "Africa/Khartoum",
	           "Africa/Mogadishu",
	           "Africa/Nairobi",
	           "Antarctica/Syowa",
	           "Asia/Aden",
	           "Asia/Amman",
	           "Asia/Baghdad",
	           "Asia/Bahrain",
	           "Asia/Beirut",
	           "Asia/Damascus",
	           "Asia/Gaza",
	           "Asia/Hebron",
	           "Asia/Istanbul",
	           "Asia/Jerusalem",
	           "Asia/Kuwait",
	           "Asia/Nicosia",
	           "Asia/Qatar",
	           "Asia/Riyadh",
	           "Asia/Tel_Aviv",
	           "EET",
	           "Etc/GMT-3",
	           "Europe/Athens",
	           "Europe/Bucharest",
	           "Europe/Chisinau",
	           "Europe/Helsinki",
	           "Europe/Istanbul",
	           "Europe/Kiev",
	           "Europe/Mariehamn",
	           "Europe/Minsk",
	           "Europe/Moscow",
	           "Europe/Nicosia",
	           "Europe/Riga",
	           "Europe/Simferopol",
	           "Europe/Sofia",
	           "Europe/Tallinn",
	           "Europe/Tiraspol",
	           "Europe/Uzhgorod",
	           "Europe/Vilnius",
	           "Europe/Volgograd",
	           "Europe/Zaporozhye",
	           "Indian/Antananarivo",
	           "Indian/Comoro",
	           "Indian/Mayotte",
	           "Israel",
	           "Turkey",
	           "W-SU",
	           "Asia/Dubai",
	           "Asia/Muscat",
	           "Asia/Tbilisi",
	           "Asia/Yerevan",
	           "Etc/GMT-4",
	           "Europe/Samara",
	           "Indian/Mahe",
	           "Indian/Mauritius",
	           "Indian/Reunion",
	           "Asia/Kabul",
	           "Asia/Tehran",
	           "Iran",
	           "Antarctica/Mawson",
	           "Asia/Aqtau",
	           "Asia/Aqtobe",
	           "Asia/Ashgabat",
	           "Asia/Ashkhabad",
	           "Asia/Baku",
	           "Asia/Dushanbe",
	           "Asia/Karachi",
	           "Asia/Oral",
	           "Asia/Samarkand",
	           "Asia/Tashkent",
	           "Asia/Yekaterinburg",
	           "Etc/GMT-5",
	           "Indian/Kerguelen",
	           "Indian/Maldives",
	           "Asia/Calcutta",
	           "Asia/Colombo",
	           "Asia/Kolkata",
	           "Asia/Kathmandu",
	           "Asia/Katmandu",
	           "Antarctica/Vostok",
	           "Asia/Almaty",
	           "Asia/Bishkek",
	           "Asia/Dacca",
	           "Asia/Dhaka",
	           "Asia/Kashgar",
	           "Asia/Novosibirsk",
	           "Asia/Omsk",
	           "Asia/Qyzylorda",
	           "Asia/Thimbu",
	           "Asia/Thimphu",
	           "Asia/Urumqi",
	           "Etc/GMT-6",
	           "Indian/Chagos",
	           "Asia/Rangoon",
	           "Indian/Cocos",
	           "Antarctica/Davis",
	           "Asia/Bangkok",
	           "Asia/Ho_Chi_Minh",
	           "Asia/Hovd",
	           "Asia/Jakarta",
	           "Asia/Krasnoyarsk",
	           "Asia/Novokuznetsk",
	           "Asia/Phnom_Penh",
	           "Asia/Pontianak",
	           "Asia/Saigon",
	           "Asia/Vientiane",
	           "Etc/GMT-7",
	           "Indian/Christmas",
	           "Antarctica/Casey",
	           "Asia/Brunei",
	           "Asia/Chita",
	           "Asia/Choibalsan",
	           "Asia/Chongqing",
	           "Asia/Chungking",
	           "Asia/Harbin",
	           "Asia/Hong_Kong",
	           "Asia/Irkutsk",
	           "Asia/Kuala_Lumpur",
	           "Asia/Kuching",
	           "Asia/Macao",
	           "Asia/Macau",
	           "Asia/Makassar",
	           "Asia/Manila",
	           "Asia/Shanghai",
	           "Asia/Singapore",
	           "Asia/Taipei",
	           "Asia/Ujung_Pandang",
	           "Asia/Ulaanbaatar",
	           "Asia/Ulan_Bator",
	           "Australia/Perth",
	           "Australia/West",
	           "Etc/GMT-8",
	           "Hongkong",
	           "PRC",
	           "ROC",
	           "Singapore",
	           "Australia/Eucla",
	           "Asia/Dili",
	           "Asia/Jayapura",
	           "Asia/Khandyga",
	           "Asia/Pyongyang",
	           "Asia/Seoul",
	           "Asia/Tokyo",
	           "Asia/Yakutsk",
	           "Etc/GMT-9",
	           "Japan",
	           "Pacific/Palau",
	           "ROK",
	           "Australia/Adelaide ",
	           "Australia/Broken_Hill",
	           "Australia/Darwin",
	           "Australia/North",
	           "Australia/South",
	           "Australia/Yancowinna ",
	           "Antarctica/DumontDUrville",
	           "Asia/Magadan",
	           "Asia/Sakhalin",
	           "Asia/Ust-Nera",
	           "Asia/Vladivostok",
	           "Australia/ACT",
	           "Australia/Brisbane",
	           "Australia/Canberra",
	           "Australia/Currie",
	           "Australia/Hobart",
	           "Australia/Lindeman",
	           "Australia/Melbourne",
	           "Australia/NSW",
	           "Australia/Queensland",
	           "Australia/Sydney",
	           "Australia/Tasmania",
	           "Australia/Victoria",
	           "Etc/GMT-10",
	           "Pacific/Chuuk",
	           "Pacific/Guam",
	           "Pacific/Port_Moresby",
	           "Pacific/Saipan",
	           "Pacific/Truk",
	           "Pacific/Yap",
	           "Australia/LHI",
	           "Australia/Lord_Howe",
	           "Antarctica/Macquarie",
	           "Asia/Srednekolymsk",
	           "Etc/GMT-11",
	           "Pacific/Bougainville",
	           "Pacific/Efate",
	           "Pacific/Guadalcanal",
	           "Pacific/Kosrae",
	           "Pacific/Noumea",
	           "Pacific/Pohnpei",
	           "Pacific/Ponape",
	           "Pacific/Norfolk",
	           "Antarctica/McMurdo",
	           "Antarctica/South_Pole",
	           "Asia/Anadyr",
	           "Asia/Kamchatka",
	           "Etc/GMT-12",
	           "Kwajalein",
	           "NZ",
	           "Pacific/Auckland",
	           "Pacific/Fiji",
	           "Pacific/Funafuti",
	           "Pacific/Kwajalein",
	           "Pacific/Majuro",
	           "Pacific/Nauru",
	           "Pacific/Tarawa",
	           "Pacific/Wake",
	           "Pacific/Wallis",
	           "NZ-CHAT",
	           "Pacific/Chatham",
	           "Etc/GMT-13",
	           "Pacific/Apia",
	           "Pacific/Enderbury",
	           "Pacific/Fakaofo",
	           "Pacific/Tongatapu",
	           "Etc/GMT-14",
	           "Pacific/Kiritimati"};

	public static Date getDateWithZoneOffset(Date dateTime, TimeZone timeZone) {
		ZoneId zoneId = timeZone.toZoneId();
		ZonedDateTime zonedDateTime = getPolicyZonedDateTime(dateTime, zoneId);

		return Date.from(zonedDateTime.toInstant());
	}

	static ZonedDateTime getPolicyZonedDateTime(Date dateTime, ZoneId policyZone) {
		ZoneOffset offsetForPolicyZone = policyZone.getRules().getOffset(dateTime.toInstant());
		ZoneOffset offsetForSystemZone = ZoneId.systemDefault().getRules().getOffset(dateTime.toInstant());

		long epochGMTSeconds = dateTime.getTime() / 1000 + offsetForSystemZone.getTotalSeconds()
				- offsetForPolicyZone.getTotalSeconds();
		Instant instant = Instant.ofEpochSecond((epochGMTSeconds));

		return ZonedDateTime.ofInstant(instant, policyZone);
	}

	public static String convertDateToString(Date date) {
		SimpleDateFormat sdf = new SimpleDateFormat(DATE_FORMAT);

		return sdf.format(date);
	}

	public static String convertTimeToString(Date date) {
		SimpleDateFormat sdf = new SimpleDateFormat(TIME_FORMAT);

		return sdf.format(date);
	}

	public static String convertDateTimeToString(Date date) {
		SimpleDateFormat sdf = new SimpleDateFormat(DATE_TIME_FORMAT);
		return sdf.format(date);
	}

	static String convertIntToDayOfWeek(int day) {
		switch (day) {
		case 1:
			return "MON";
		case 2:
			return "TUE";
		case 3:
			return "WED";
		case 4:
			return "THU";
		case 5:
			return "FRI";
		case 6:
			return "SAT";
		case 7:
			return "SUN";
		default:
			return null;
		}
	}

}
