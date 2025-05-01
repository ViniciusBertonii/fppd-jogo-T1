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

func verificarDesativarArmadilha(jogo *Jogo) {
	direcoes := []struct{ dx, dy int }{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0}, // cima, direita, baixo, esquerda
	}

	for _, dir := range direcoes {
		nx := jogo.PosX + dir.dx
		ny := jogo.PosY + dir.dy

		if ny >= 0 && ny < len(jogo.Mapa) && nx >= 0 && nx < len(jogo.Mapa[0]) {
			if jogo.Mapa[ny][nx].simbolo == Armadilha.simbolo {
				// Achou uma armadilha perto → desativa
				select {
				case canalArmadilha <- true: // envia sinal para desativar
					jogo.StatusMsg = "🔵 Você desativou uma armadilha!"
				default:
					// evita travar caso o canal esteja fechado
				}
			}
		}
	}
}


// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		personagemInteragir(jogo) // interação normal
		verificarDesativarArmadilha(jogo) // adicional: tenta desativar armadilha próxima
	
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
