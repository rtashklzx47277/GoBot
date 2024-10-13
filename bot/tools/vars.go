package tools

var (
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.0.0"
	ClientVersion = "2.20240620.05.00"
	DefaultImage  = "https://i.imgur.com/2wAkxNb.png"

	ChannelList = map[string]string{
		"UCJFZiqLMntJufDCHc6bQixg": "ホロライブ",
		"UCp6993wxpyDPHUpavwDFqgg": "ときのそら",
		"UCDqI2jOz0weumE8s7paEk6g": "ロボ子さん",
		"UC5CwaMl1eIgY8h02uZw7u8A": "星街すいせい",
		"UC-hM6YJuNYVAmUWxeIr9FeA": "さくらみこ",
		"UC0TXe_LYZ4scaW2XMyi5_kw": "AZKi",
		"UCD8HOxPs4Xvsm8H0ZxXGiBw": "夜空メル",
		"UCdn5BQ06XqgXoAxIhbqw5Rg": "白上フブキ",
		"UCQ0UDLQCjY0rmuxCDE38FGg": "夏色まつり",
		"UCFTLzh12_nrtzqBPsTCqenA": "アキ・ローゼンタール",
		"UC1CfXB_kRs3C-zaeTG3oGyg": "赤井はあと",
		"UC1opHUrw8rvnsadT-iGp7Cg": "湊あくあ",
		"UCXTpFs_3PqI41qX2d9tL2Rw": "紫咲シオン",
		"UC7fk0CB07ly8oSl0aqKkqFg": "百鬼あやめ",
		"UC1suqwovbL1kzsoaZgFZLKg": "癒月ちょこ",
		"UCvzGlP9oQwU--Y0r9id_jnA": "大空スバル",
		"UCp-5t9SrOQwXMU7iIjQfARg": "大神ミオ",
		"UCvaTdHTWBGv3MKj3KVqJVCw": "猫又おかゆ",
		"UChAnqc_AY5_I3Px5dig3X1Q": "戌神ころね",
		"UC1DCedRgGHBdm81E1llLhOQ": "兎田ぺこら",
		"UCl_gCybOJRIgOXw6Qb4qJzQ": "潤羽るしあ",
		"UCvInZx9h3jC2JzsIzoOebWg": "不知火フレア",
		"UCdyqAaZDKHXg4Ahi7VENThQ": "白銀ノエル",
		"UCCzUftO8KOVkV4wQG1vkUvg": "宝鐘マリン",
		"UCZlDXzGoo7d44bwdNObFacg": "天音かなた",
		"UCS9uQI-jC3DE0L4IpXyvr6w": "桐生ココ",
		"UCqm3BQLlJfvkTsX_hvm0UmA": "角巻わため",
		"UC1uv2Oq6kNxgATlCiez59hw": "常闇トワ",
		"UCa9Y57gfeY0Zro_noHRVrnw": "姫森ルーナ",
		"UCFKOVgVbGmX65RxO3EtH3iw": "雪花ラミィ",
		"UCAWSyEs_Io8MtpY3m-zqILA": "桃鈴ねね",
		"UCUKD-uaobj9jiqB-VXt71mA": "獅白ぼたん",
		"UCgZuwn-O7Szh9cAgHqJ6vjw": "魔乃アロエ",
		"UCK9V2B22uJYu3N7eR_BT9QA": "尾丸ポルカ",
		"UCENwRMx5Yh42zWpzURebzTw": "ラプラス・ダークネス",
		"UCs9_O1tRPMQTHQ-N_L6FU2g": "鷹嶺ルイ",
		"UC6eWCld0KwmyHFbAqK3V-Rw": "博衣こより",
		"UCIBY1ollUsauvVi4hW4cumw": "沙花叉クロヱ",
		"UC_vMYWcDjmfdpH6r4TTn1MQ": "風真いろは",
		"UCMGfV7TVTmHhEErVJg1oHBQ": "火威青",
		"UCWQtYtq9EOB4-I5P-3fh8lA": "音乃瀬奏",
		"UCtyWhCj3AqKh2dXctLkDtng": "一条莉々華",
		"UCdXAk5MpyLD8594lm_OvtGQ": "儒烏風亭らでん",
		"UC1iA6_NT4mtAcIII6ygrvCw": "轟はじめ",
		"UCOyYb1c43VlX9rc_lT6NKQw": "Ayunda Risu",
		"UCP0BspO_AMEe3aQqqpo89Dg": "Moona Hoshinova",
		"UCAoy6rzhSf4ydcYjJw3WoVg": "Airani Iofifteen",
		"UCYz_5n-uDuChHtLo7My1HnQ": "Kureiji Ollie",
		"UC727SQYUvx5pDDGQpTICNWg": "Anya Melfissa",
		"UChgTyjG-pdNvxxhdsXfHQ5Q": "Pavolia Reine",
		"UCTvHWSfBZgtxE4sILOaurIQ": "Vestia Zeta",
		"UCZLZ8Jjx_RN2CXloOmgTHVg": "Kaela Kovalskia",
		"UCjLEmnpCNeisMxy134KPwWw": "Kobo Kanaeru",
		"UCL_qhgtOy0dy1Agp8vkySQg": "Mori Calliope",
		"UCHsx4Hqa-1ORjQTh9TYDhww": "Takanashi Kiara",
		"UCMwGHR0BTZuLsmjY_NT5Pwg": "Ninomae Ina'nis",
		"UCoSrY_IQQVpmIRZ9Xf-y93g": "Gawr Gura",
		"UCyl1z3jo3XHR1riLFKG5UAg": "Watson Amelia",
		"UCsUj0dszADCGbF3gNrQEuSQ": "Tsukumo Sana",
		"UCO_aKKYxn4tvrqPjcTzZ6EQ": "Ceres Fauna",
		"UCmbs8T6MWqUHP1tIQvSgKrg": "Ouro Kronii",
		"UC3n5uGu18FoCy23ggWWp8tA": "Nanashi Mumei",
		"UCgmPnx-EEeOrZSg5Tiw7ZRQ": "Hakos Baelz",
		"UC8rcEBzJSleTkf_-agPM20g": "IRyS",
		"UC9p_lqQ0FEDz327Vgf5JwqA": "Koseki Bijou",
		"UCgnfPPb9JI3e9A4cXHnWbyg": "Shiori Novella",
		"UC_sFNM0z0MWm9A6WlKPuMMg": "Nerissa Ravencroft",
		"UCt9H_RpQzhxzlyBxFqrdHqA": "FUWAMOCO",
		"UCW5uhrG1eCBYditmhL0Ykjw": "Elizabeth Rose Bloodflame",
		"UCDHABijvPBnJm7F-KlNME3w": "Gigi Murin",
		"UCvN5h1ShZtc7nly3pezRayg": "Cecilia Immergreen",
		"UCl69AEx4MdqMZH7Jtsm7Tig": "Raora Panthera",
		"UCWCc8tO-uUl_7SJXIKJACMw": "神楽めあ",
		"UC8NZiqKx6fsDT3AVcMiVFyA": "犬山たまき",
		"UC_4tXjqecqox5Uc05ncxpxg": "椎名唯華",
		"UCoztvTULBYd3WmStqYeoHcA": "笹木咲",
		"UC9V3Y3_uzU5e-usObb6IE1w": "星川サラ",
		"UC9EjSJ8pvxtvPdxLOElv73w": "魔界ノりりむ",
	}

	UserData = map[string]map[string]map[string]string{
		"Aqua": {
			"Youtube": {
				"Id":               "UC1opHUrw8rvnsadT-iGp7Cg",
				"DiscordChannelId": "965968317280055397",
			},
			"Twitch": {
				"Id":               "738746247",
				"DiscordChannelId": "970990916980584508",
			},
			"Twitcasting": {
				"Id":               "1024528894940987392",
				"DiscordChannelId": "969892552486563880",
			},
			"Tiktok": {
				"Id":               "minatoaqua_hololive",
				"DiscordChannelId": "969862812333666334",
			},
			"Fanbox": {
				"Id":               "80355000",
				"DiscordChannelId": "965967553870594098",
			},
			"News": {
				"Id":               "湊あくあ",
				"DiscordChannelId": "968838661569400952",
			},
		},
		"Shion": {
			"Youtube": {
				"Id":               "UCXTpFs_3PqI41qX2d9tL2Rw",
				"DiscordChannelId": "965973309432942642",
			},
			"Twitch": {
				"Id":               "773041510",
				"DiscordChannelId": "976847607248871444",
			},
			"Twitcasting": {
				"Id":               "1024533638879166464",
				"DiscordChannelId": "976850346292961310",
			},
			"Tiktok": {
				"Id":               "murasakishion_hololive",
				"DiscordChannelId": "996703905612320778",
			},
			"Fanbox": {
				"Id":               "69014608",
				"DiscordChannelId": "872814425046937610",
			},
			"News": {
				"Id":               "紫咲シオン",
				"DiscordChannelId": "971271623221075968",
			},
		},
		"Rinchan": {
			"Youtube": {
				"Id":               "UCrV1Hf5r8P148idjoSfrGEQ",
				"DiscordChannelId": "965967553870594098",
			},
			"Fanbox": {
				"Id":               "80355000",
				"DiscordChannelId": "965967553870594098",
			},
		},
		"Rinco": {
			"Youtube": {
				"Id":               "UCBbGcCpI1NmpNdEV1qPQttw",
				"DiscordChannelId": "965967553870594098",
			},
		},
		"DesuRinco": {
			"Youtube": {
				"Id":               "UC0G3JPhTMpZh1r-TvUbimzA",
				"DiscordChannelId": "965967553870594098",
			},
		},
	}
)
