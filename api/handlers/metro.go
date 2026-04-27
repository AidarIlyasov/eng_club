package handlers

import (
	"github.com/valyala/fasthttp"
)

type MetroStation struct {
	Line  int    `json:"line"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

// GetMetroList handles GET /api/metro-list.
func (d *Deps) GetMetroList(ctx *fasthttp.RequestCtx) {
	stations := []MetroStation{
		{Line: 1, Value: "avtovo", Name: "Avtovo"},
		{Line: 5, Value: "admiraltejskaya", Name: "Admiraltejskaya"},
		{Line: 1, Value: "akademicheskaya", Name: "Akademicheskaya"},
		{Line: 1, Value: "baltijskaya", Name: "Baltijskaya"},
		{Line: 3, Value: "begovaya", Name: "Begovaya"},
		{Line: 5, Value: "buharestskaya", Name: "Buharestskaya"},
		{Line: 3, Value: "vasileostrovskaya", Name: "Vasileostrovskaya"},
		{Line: 1, Value: "vladimirskaya", Name: "Vladimirskaya"},
		{Line: 5, Value: "volkovskaya", Name: "Volkovskaya"},
		{Line: 1, Value: "vyborgskaya", Name: "Vyborgskaya"},
		{Line: 2, Value: "gorkovskaya", Name: "Gorkovskaya"},
		{Line: 3, Value: "gostinyj-dvor", Name: "Gostinyj Dvor"},
		{Line: 1, Value: "grazhdanskij-prospekt", Name: "Grazhdanskij Prospekt"},
		{Line: 1, Value: "devyatkino", Name: "Devyatkino"},
		{Line: 4, Value: "dostoevskaya", Name: "Dostoevskaya"},
		{Line: 5, Value: "dunajskaya", Name: "Dunajskaya"},
		{Line: 3, Value: "elizarovskaya", Name: "Elizarovskaya"},
		{Line: 5, Value: "zvenigorodskaya", Name: "Zvenigorodskaya"},
		{Line: 2, Value: "zvezdnaya", Name: "Zvezdnaya"},
		{Line: 1, Value: "kirovskij-zavod", Name: "Kirovskij Zavod"},
		{Line: 5, Value: "komendantskij-prospekt", Name: "Komendantskij Prospekt"},
		{Line: 5, Value: "krestovskij-ostrov", Name: "Krestovskij Ostrov"},
		{Line: 2, Value: "kupchino", Name: "Kupchino"},
		{Line: 4, Value: "ladozhskaya", Name: "Ladozhskaya"},
		{Line: 1, Value: "leninskij-prospekt", Name: "Leninskij Prospekt"},
		{Line: 1, Value: "lesnaya", Name: "Lesnaya"},
		{Line: 4, Value: "ligovskij-prospekt", Name: "Ligovskij Prospekt"},
		{Line: 3, Value: "lomonosovskaya", Name: "Lomonosovskaya"},
		{Line: 3, Value: "mayakovskaya", Name: "Mayakovskaya"},
		{Line: 5, Value: "mezhdunarodnaya", Name: "Mezhdunarodnaya"},
		{Line: 2, Value: "moskovskaya", Name: "Moskovskaya"},
		{Line: 2, Value: "moskovskie-vorota", Name: "Moskovskie Vorota"},
		{Line: 1, Value: "narvskaya", Name: "Narvskaya"},
		{Line: 2, Value: "nevskij-prospekt", Name: "Nevskij Prospekt"},
		{Line: 4, Value: "novocherkasskaya", Name: "Novocherkasskaya"},
		{Line: 3, Value: "obuhovo", Name: "Obuhovo"},
		{Line: 5, Value: "obvodnyj-kanal", Name: "Obvodnyj Kanal"},
		{Line: 2, Value: "ozerki", Name: "Ozerki"},
		{Line: 2, Value: "park-pobedy", Name: "Park Pobedy"},
		{Line: 2, Value: "parnas", Name: "Parnas"},
		{Line: 2, Value: "petrogradskaya", Name: "Petrogradskaya"},
		{Line: 2, Value: "pionerskaya", Name: "Pionerskaya"},
		{Line: 3, Value: "ploshchad-aleksandra-nevskogo-1", Name: "Ploshchad Aleksandra Nevskogo-1"},
		{Line: 4, Value: "ploshchad-aleksandra-nevskogo-2", Name: "Ploshchad Aleksandra Nevskogo-2"},
		{Line: 1, Value: "ploshchad-vosstaniya", Name: "Ploshchad Vosstaniya"},
		{Line: 1, Value: "ploshchad-lenina", Name: "Ploshchad Lenina"},
		{Line: 1, Value: "ploshchad-muzhestva", Name: "Ploshchad Muzhestva"},
		{Line: 1, Value: "politekhnicheskaya", Name: "Politekhnicheskaya"},
		{Line: 3, Value: "primorskaya", Name: "Primorskaya"},
		{Line: 3, Value: "proletarskaya", Name: "Proletarskaya"},
		{Line: 4, Value: "prospekt-bolshevikov", Name: "Prospekt Bolshevikov"},
		{Line: 5, Value: "prospekt-slavy", Name: "Prospekt Slavy"},
		{Line: 1, Value: "prospekt-veteranov", Name: "Prospekt Veteranov"},
		{Line: 2, Value: "prospekt-prosveshcheniya", Name: "Prospekt Prosveshcheniya"},
		{Line: 1, Value: "pushkinskaya", Name: "Pushkinskaya"},
		{Line: 3, Value: "rybackoe", Name: "Rybackoe"},
		{Line: 5, Value: "sadovaya", Name: "Sadovaya"},
		{Line: 2, Value: "sennaya", Name: "Sennaya Ploshchad"},
		{Line: 5, Value: "shushary", Name: "Shushary"},
		{Line: 4, Value: "spasskaya", Name: "Spasskaya"},
		{Line: 5, Value: "sportivnaya", Name: "Sportivnaya"},
		{Line: 5, Value: "staraya-derevnya", Name: "Staraya Derevnya"},
		{Line: 1, Value: "tekhnologicheskij-institut-1", Name: "Tekhnologicheskij Institut-1"},
		{Line: 2, Value: "tekhnologicheskij-institut-2", Name: "Tekhnologicheskij Institut-2"},
		{Line: 2, Value: "udelnaya", Name: "Udelnaya"},
		{Line: 4, Value: "ulica-dybenko", Name: "Uica Dybenko"},
		{Line: 2, Value: "frunzenskaya", Name: "Frunzenskaya"},
		{Line: 2, Value: "chernaya-rechka", Name: "Chernaya Rechka"},
		{Line: 1, Value: "chernyshevskaya", Name: "Chernyshevskaya"},
		{Line: 5, Value: "chkalovskaya", Name: "Chkalovskaya"},
		{Line: 2, Value: "ehlektrosila", Name: "Ehlektrosila"},
	}

	WriteJSON(ctx, fasthttp.StatusOK, stations)
}