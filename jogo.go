// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main //novas funcionalidades

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
	StartX, StartY int // posição inicial salva
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
				jogo.PosX, jogo.PosY = x, y
				jogo.StartX, jogo.StartY = x, y // <<< NOVO: salva o ponto inicial também
			
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

func adicionarArmadilhasAleatorias(jogo *Jogo, quantidade int) {
	for i := 0; i < quantidade; i++ {
		x, y := encontrarPosicaoLivre(jogo)

		go iniciarArmadilhaAleatoria(jogo, x, y)
	}
}

func iniciarArmadilhaAleatoria(jogo *Jogo, x, y int) {
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
		default:
			time.Sleep(100 * time.Millisecond) // para evitar loop intenso
		}

		// Verifica se jogador caiu na armadilha ativa
		jogo.Mutex.Lock()
		if ativa && jogo.PosX == x && jogo.PosY == y {
			// Volta o jogador para o ponto inicial
			jogo.StatusMsg = "💥 Você caiu numa armadilha! Voltando para o início!"
			jogo.PosX, jogo.PosY = encontrarPontoInicial(jogo)
		}
		jogo.Mutex.Unlock()
	}
}

func encontrarPontoInicial(jogo *Jogo) (int, int) {
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			if elem.simbolo == Personagem.simbolo {
				return x, y
			}
		}
	}
	// fallback caso não encontre (não deveria acontecer)
	return 1, 1
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

func movimentarInimigoVertical(jogo *Jogo, xInicial, yInicial int) {
	x, y := xInicial, yInicial
	direcao := 1

	for {
		time.Sleep(500 * time.Millisecond)

		jogo.Mutex.Lock()

		// Verifica se o inimigo ainda existe
		if jogo.Mapa[y][x].simbolo != Inimigo.simbolo {
			jogo.Mutex.Unlock()
			return
		}

		ny := y + direcao

		// Verifica limites ou parede
		if ny < 0 || ny >= len(jogo.Mapa) || jogo.Mapa[ny][x].tangivel {
			direcao *= -1
			jogo.Mutex.Unlock()
			continue
		}

		// Verifica se o jogador está na próxima posição
		if jogo.PosX == x && jogo.PosY == ny {
			// Jogador está no caminho → inverte a direção
			direcao *= -1
			jogo.StatusMsg = "⚡ O inimigo tentou te atingir, mas mudou de direção!"
			jogo.Mutex.Unlock()
			continue
		}

		// Verifica se o destino é Vazio ou Vegetação
		destino := jogo.Mapa[ny][x].simbolo
		if destino == Vazio.simbolo || destino == Vegetacao.simbolo {
			// Apaga a posição antiga
			jogo.Mapa[y][x] = Vazio
			// Move para a nova posição
			jogo.Mapa[ny][x] = Inimigo
			y = ny
		} else {
			// Se não puder andar, inverte
			direcao *= -1
		}

		jogo.Mutex.Unlock()

		interfaceDesenharJogo(jogo)
	}
}





func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}
	// Se for armadilha, pode pisar, mesmo que tangível
	if jogo.Mapa[y][x].simbolo == Armadilha.simbolo {
		return true
	}
	// Senão, verifica normalmente
	if jogo.Mapa[y][x].tangivel {
		return false
	}
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
