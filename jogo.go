// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"math/rand"
	"os"
	"sync"
	"time"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo   rune
	cor       Cor
	corFundo  Cor
	tangivel  bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa            [][]Elemento // grade 2D representando o mapa
	PosX, PosY      int          // posição atual do personagem
	UltimoVisitado  Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg       string       // mensagem para a barra de status
	sync.Mutex                   // adiciona o mutex para sincronização
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
	Portal 	   = Elemento{'○', CorVerde, CorPadrao, false}
	Armadilha = Elemento{'▲', CorVermelho, CorPadrao, true}

)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio}
}




// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()

	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				e = Inimigo
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func iniciarSentinela(jogo *Jogo, xInicial, y, xFinal int) {
	x := xInicial
	direcao := 1

	for {
		jogo.Mutex.Lock()

		// Verifica se encostou no personagem
		if jogo.PosX == x && jogo.PosY == y {
			jogo.StatusMsg = "⚠️  O inimigo patrulheiro te pegou! Cuidado!"
		}

		// Remove inimigo da posição anterior
		if jogo.Mapa[y][x].simbolo == Inimigo.simbolo {
			jogo.Mapa[y][x] = Vazio
		}

		// Calcula nova posição
		nx := x + direcao
		if nx < xInicial || nx > xFinal || jogo.Mapa[y][nx].tangivel {
			direcao *= -1
			nx = x + direcao
		}

		// Coloca inimigo na nova posição
		if jogo.Mapa[y][nx].simbolo == Vazio.simbolo {
			jogo.Mapa[y][nx] = Inimigo
		}

		x = nx
		jogo.Mutex.Unlock()

		interfaceDesenharJogo(jogo)
		time.Sleep(400 * time.Millisecond)
	}
}

func iniciarArmadilha(jogo *Jogo, x, y int, canalDesativar <-chan bool) {
	ativa := false

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ativa = !ativa
			jogo.Mutex.Lock()
			if ativa {
				jogo.Mapa[y][x] = Armadilha
			} else {
				jogo.Mapa[y][x] = Vazio
			}
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)

		case <-canalDesativar:
			jogo.Mutex.Lock()
			jogo.Mapa[y][x] = Vazio
			jogo.StatusMsg = "🔕 Armadilha foi desativada!"
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)
			return
		}

		// Verifica se jogador pisou
		jogo.Mutex.Lock()
		if ativa && jogo.PosX == x && jogo.PosY == y {
			jogo.StatusMsg = "💥 Você pisou numa armadilha!"
		}
		jogo.Mutex.Unlock()
	}
}




func encontrarPosicaoLivre(jogo *Jogo) (int, int) {
	for {
		x := rand.Intn(len(jogo.Mapa[0]))
		y := rand.Intn(len(jogo.Mapa))
		if !jogo.Mapa[y][x].tangivel && jogo.Mapa[y][x].simbolo == Vazio.simbolo {
			return x, y
		}
	}
}


func iniciarPortal(jogo *Jogo) {
	for {
		x, y := encontrarPosicaoLivre(jogo)

		jogo.Mutex.Lock()
		jogo.Mapa[y][x] = Portal
		jogo.Mutex.Unlock()

		interfaceDesenharJogo(jogo)

		// Canal de timeout com 5 segundos
		timeout := time.After(5 * time.Second)
		tick := time.Tick(300 * time.Millisecond)

	loop:
		for {
			select {
			case <-timeout:
				jogo.Mutex.Lock()
				// Remove se ainda for o portal
				if jogo.Mapa[y][x].simbolo == Portal.simbolo {
					jogo.Mapa[y][x] = Vazio
					jogo.StatusMsg = "⏱️ O portal desapareceu!"
				}
				jogo.Mutex.Unlock()
				interfaceDesenharJogo(jogo)
				break loop

			case <-tick:
				jogo.Mutex.Lock()
				if jogo.PosX == x && jogo.PosY == y {
					jogo.Mapa[y][x] = Vazio
					jogo.StatusMsg = "🚪 Você entrou no portal a tempo!"
					jogo.Mutex.Unlock()
					interfaceDesenharJogo(jogo)
					break loop
				}
				jogo.Mutex.Unlock()
			}
		}

		time.Sleep(3 * time.Second) // tempo antes do próximo portal
	}
}


// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado     // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]   // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento              // move o elemento
}
