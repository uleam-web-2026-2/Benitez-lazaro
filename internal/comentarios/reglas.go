package comentarios

// transiciones permitidas: de -> a. Solo se avanza, nunca se retrocede.
var transiciones = map[string]string{
	EstadoPendiente: EstadoAprobado,
	EstadoAprobado:  EstadoDestacado,
}

// TransicionPermitida dice si un comentario puede pasar del estado "de" al estado "a".
func TransicionPermitida(de, a string) bool {
	siguiente, ok := transiciones[de]
	return ok && siguiente == a
}

// MedallaPara devuelve la medalla simbólica que corresponde a una reputación.
func MedallaPara(reputacion int) string {
	switch {
	case reputacion >= 50:
		return "Expert"
	case reputacion >= 25:
		return "Ayudante"
	default:
		return "Novato"
	}
}
