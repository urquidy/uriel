package main

import (
	"strings"
	"log"
	"time"
	"os"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	conf = Configuration{}
	warning = false
)

type Configuration struct {
	Session 	*discordgo.Session
	Channel		string	`json:"channel"`
	BotID 		string 
	Token		string 	`json:"token"`
}

func main() {
	var err error

	//Configuration Loader
	handler, err := os.Open("uriel_config.json")
	if err != nil {
		log.Println("No configuration file found.")
		log.Panic(err)
	}

	err = json.NewDecoder(handler).Decode(&conf)
	check(err)

	err = handler.Close()
	check(err)
	// ----------------------------

	conf.Session, err = discordgo.New("Bot " + conf.Token)
	check(err)
	
	u, err := conf.Session.User("@me")
	check(err)
	
	conf.BotID = u.ID
	conf.Session.AddHandler(ChatMonitor)

	err = conf.Session.Open()
	check(err)

	log.Println("Uriel logged in.")

	Updater()
	
	<-make(chan struct{})
	return
}

// Commands
func cmdHelp(args []string) {
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"Comandos Uriel:\n**!hora** : Muestra la hora de España y del host\n**!noticias** : Muestra las noticias mas recientes de la página de Metin2.es\n**!eventos** : Muestra los eventos fijos del juego")
}

func cmdHour(args []string) {
	t := CEST()
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"Hora actual Venezuela : " + time.Now().Format(time.RFC850) + "\nHora actual España : " + t.Full)
}

func cmdNews(args []string) {
	resp, err := http.Get("https://board.es.metin2.gameforge.com/")
	check(err)

	root, err := html.Parse(resp.Body)
	check(err)

	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Header || n.DataAtom == atom.Div {
			return scrape.Attr(n,"class") == "messageHeader" || scrape.Attr(n,"class") == "messageText"
		}
		return false
	}

	articles := scrape.FindAll(root, matcher)
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**Noticias Metin2.es**")
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**" + scrape.Text(articles[0]) + "**" + "\n" + scrape.Text(articles[1]))
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**" + scrape.Text(articles[2]) + "**" + "\n" + scrape.Text(articles[3]))
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**" + scrape.Text(articles[4]) + "**" + "\n" + scrape.Text(articles[5]))
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**" + scrape.Text(articles[6]) + "**" + "\n" + scrape.Text(articles[7]))
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**" + scrape.Text(articles[8]) + "**" + "\n" + scrape.Text(articles[9]))
}

func cmdEvents(args []string) {
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"**Eventos Fijos Metin2.es**")
	_, _ = conf.Session.ChannelMessageSend(conf.Channel,"\n**Lunes:**\n\t22:00-00:00 Drop de supermonturas\n\n**Martes:**\n\t16:00-18:00 Drop de supermonturas\n\n**Miércoles:**\n\t23:00-01:00 Drop de supermonturas\n\n**Jueves:**\n\t13:00-15:00 Drop de supermonturas\n\n**Viernes:**\n\t18:00-20:00 Drop de supermonturas\n\n**Sábado:**\n\t13:00-15:00 Drop de supermonturas\n\t18:00-00:00 (El segundo sábado de cada mes) Drop de cajas luz luna\n\tTodo el día (Último fin de semana de cada mes) Festival de la cosecha\n\n**Domingo:**\n\t13:00-15:00 Drop de supermonturas\n\t22:00-02:00 Drop de Alubias verdes del dragón\n\tTodo el día (Último fin de semana de cada mes) Festival de la cosecha\n\n**Cualquier día de la semana:**\n\tCompetición OX (Mínimo una vez por semana)\n\tDrop de telas delicadas 4 h. (Sin día ni hora predefinido, sale el mensaje en el juego cuando comienza)\n\tDrop de Cor Draconis 4 h. (Sin día ni hora predefinido, sale el mensaje en el juego cuando comienza).")
}
//------------

func LoadCommands() map[string]func([]string) {
	return map[string]func([]string){
		"!ayuda":cmdHelp,
		"!hora":cmdHour,
		"!noticias":cmdNews,
		"!eventos":cmdEvents,
	}
}

func ChatMonitor(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == conf.BotID {
		return
	}

	if m.ChannelID != conf.Channel {
		return
	}

	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	var cmd map[string]func([]string)
	cmd = make(map[string]func([]string))
	cmd = LoadCommands()

	input := strings.Split(m.Content," ")

	command := strings.ToLower(input[0])

	val, exist := cmd[command]
	if exist {
		val(input)
	} else {
		_, _ = conf.Session.ChannelMessageSend(conf.Channel,"Comando inválido. Usa **!ayuda** para ver la lista de comandos disponibles")
	}


}

type TimeEx struct {
	Hour		int
	Min			int
	Sec			int
	Day			int
	Weekday		string
	Month		string
	Full	 	string
	Year		int
}

func CEST()	TimeEx {
	t := TimeEx{}

	locat, err := time.LoadLocation("Europe/Madrid")
	check(err)

	cest := time.Now().In(locat)
	var vemonth time.Month
	t.Year, vemonth, t.Day = cest.Date()
	t.Month = vemonth.String()
	t.Hour, t.Min, t.Sec = cest.Clock()
	t.Weekday = cest.Weekday().String()
	t.Full = cest.Format(time.RFC850)

	return t
}

func Updater() {
	t := CEST()

	switch t.Weekday {
		case "Monday":
			if t.Hour >= 22 && t.Hour < 24 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true
				}
			} else if warning == true {
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}

		case "Tuesday":
			if t.Hour >= 16 && t.Hour < 18 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true				
				}
			} else if warning == true {
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}

		case "Wednesday":
			if t.Hour >= 23 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true				
				}
			}

		case "Thursday":
			if t.Hour >= 13 && t.Hour < 15 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true				
				}
			} else if warning == true {
				if t.Hour > 1 && t.Hour <= 2 {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
					warning = false
					return
				}
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}

		case "Friday":
			if t.Hour >= 18 && t.Hour < 20 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true				
				}
			} else if warning == true {
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}

		case "Saturday":
			if t.Hour >= 13 && t.Hour < 15 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl Festival de la Cosecha ha comenzado.\nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true
				}
			} else if warning == true {
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}

		case "Sunday":
			if t.Hour >= 13 && t.Hour < 15 {
				if warning == false {
					_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha empezado y terminará en dos horas. Hora actual CEST: " + t.Full)
					warning = true				
				}
			} else if warning == true {
				_, _ = conf.Session.ChannelMessageSend(conf.Channel,"@everyone \nEl drop de supermonturas ha terminado. Hora actual CEST: " + t.Full)
				warning = false
			}
	}

	_ = time.AfterFunc(5 * time.Minute, Updater)
}

func check(e error) {
	if e != nil {
		log.Panic(e)
	}
}
