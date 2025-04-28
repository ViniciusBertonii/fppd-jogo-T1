// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy

	if jogoPodeMoverPara(jogo, nx, ny) {
		elementoDestino := jogo.Mapa[ny][nx] // <<< pega o elemento ANTES de mover!

		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny

		// Agora sim, checar se caiu na armadilha
		if elementoDestino.simbolo == Armadilha.simbolo {
			jogo.StatusMsg = "💥 Você caiu numa armadilha!"
			jogo.PosX, jogo.PosY = jogo.StartX, jogo.StartY
		}
	}
}


// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
func personagemInteragir(jogo *Jogo) {
	// Atualmente apenas exibe uma mensagem de status
	jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
