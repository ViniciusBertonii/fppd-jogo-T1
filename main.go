// main.go - Loop principal do jogo
package main //novas funcionalidades

import (
	"os"
	"sync"
)

// Mensagem é usada para comunicação entre goroutines
type Mensagem struct {
	Tipo string
	Dados any
}

// Canais globais para comunicação entre goroutines
var (
	CanalAlertaInimigo = make(chan Mensagem)
	CanalPortal        = make(chan Mensagem)
	CanalArmadilha     = make(chan Mensagem)
)

var canalArmadilha chan bool


func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o jogo
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Desenha o estado inicial do jogo
	interfaceDesenharJogo(&jogo)

	jogo.Mutex = sync.Mutex{} // Garante que esteja inicializado
	


	//go iniciarSentinela(&jogo, 10, 5, 20)      // patrulha de x=10 até x=20 na linha 5
	go iniciarPortal(&jogo)                    // surge e some em posições aleatórias
	canalArmadilha := make(chan bool)
	go iniciarArmadilha(&jogo, 25, 15, canalArmadilha)

	canalArmadilha = make(chan bool) // inicializa o canal

	go iniciarArmadilha(&jogo, 30, 15, canalArmadilha) // armadilha fixa, como já estava

	

	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Inimigo.simbolo {
				go movimentarInimigoVertical(&jogo, x, y)
			}
		}
	}

	adicionarArmadilhasAleatorias(&jogo, 3) // por exemplo, 3 armadilhas aleatórias

	
	



	// Loop principal de entrada
	for {
		evento := interfaceLerEventoTeclado()
		if continuar := personagemExecutarAcao(evento, &jogo); !continuar {
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}