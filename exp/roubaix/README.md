# RouletTEa Roubaisienne

Un petit roguelike en Go — un Canadien se réveille à Roubaix et doit
retrouver son passeport pour rentrer au pays.

## Histoire

Vous êtes Canadien. Vous vous réveillez dans une chambre inconnue à
Roubaix, une vieille ville industrielle du nord de la France. L'air
est humide, les rues sont pavées, et personne ne parle anglais. Vous
devrez explorer la ville, descendre dans les profondeurs de la vieille
filature, et affronter les catacombes pour retrouver votre
passeport. Le sirop d'érable vous manque déjà.

## Niveaux

1. **Roubaix — le centre-ville** — ruelles pavées, cafés, canaux. Des
   chiens errants et des ivrognes rôdent.
2. **La Vieille Filature** — l'usine désaffectée. Des rats géants et
   d'étranges machines vous attendent.
3. **Les Catacombes** — sous la ville. Spectres et squelettes gardent
   les secrets oubliés.

## Items

| Objet              | Effet                          |
|--------------------|--------------------------------|
| Baguette `%`       | Soigne 3 PV                    |
| Café `[`           | Soigne 2 PV                    |
| Béret `^`          | Défense +1                     |
| Poutine `!`        | Soigne 8 PV (étage 2)          |
| Clé `$`            | Utile? (étage 2)               |
| Sirop d'érable `♥` | Soigne 15 PV (étage 3)         |
| Passeport `○`      | **GAGNER LA PARTIE** (étage 3) |

## Commandes

| Touche                | Action                                            |
|-----------------------|---------------------------------------------------|
| `h j k l` / `← ↓ ↑ →` | Déplacement                                       |
| `y u b n`             | Déplacement diagonal                              |
| `.`                   | Attendre                                          |
| `g`                   | Ramasser un objet                                 |
| `i`                   | Inventaire (appuyez sur une lettre pour utiliser) |
| `d`                   | Déposer le premier objet                          |
| `>`                   | Descendre un escalier                             |
| `<`                   | Monter un escalier                                |
| `?`                   | Aide                                              |
| `q`                   | Quitter                                           |

## Lancer

```bash
go run .
```

## Technologie

- [Bubble Tea v2](https://charm.land/bubbletea/v2) — TUI framework
- [Lip Gloss v2](https://charm.land/lipgloss/v2) — styles et couleurs
- Génération procédurale, champ de vision (FOV), combat au corps-à-corps
