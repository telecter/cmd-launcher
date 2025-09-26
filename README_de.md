<img src="docs/icon.png" width="180">

# cmd-launcher

Ein minimalisticher Minecraft Launcher für die Command Line.  
[EN](README.md) | DE

[![Build](https://github.com/telecter/cmd-launcher/actions/workflows/build.yml/badge.svg)](https://github.com/telecter/cmd-launcher/actions/workflows/build.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/telecter/cmd-launcher)
[![Go Reference](https://pkg.go.dev/badge/github.com/telecter/cmd-launcher.svg)](https://pkg.go.dev/github.com/telecter/cmd-launcher)

- [Installation](#installation)
  - [Von Source installieren](#von-source-installieren)
- [Verwendung](#verwendung)
  - [Instanze](#instanze)
  - [Das Spiel starten](#das-spiel-starten)
  - [Authentifizierung](#authentifizierung)
  - [Instanzkonfiguration](#instanzkonfiguration)
  - [Suchen](#suchen)

[API Dokumentation](docs/API.md)

## Installation

Stellen Sie erst sicher dass [Go](https://go.dev) installiert ist.

**Um die neuste Version zu installieren:**

```bash
go install github.com/telecter/cmd-launcher@latest
```

### Von Source installieren

1. Repository klonen: `git clone https://github.com/telecter/cmd-launcher`
2. In dem Sourcecode-Verzeichnis, führen Sie `go run .` aus um den Launcher zu bauen und starten.
3. Wenn Sie bereit sind, kompilieren Sie den ausführbare Datei mit `go build .`

## Verwendung

Verwenden Sie die `--help` Option für den Aufruf eines Befehles.

### Instanze

**Instanze erstellen**  
Um eine neue Instanz zu erstellen, führen Sie den `inst create` Befehl aus.  
Sie können auch die `--loader, -l` Option verwenden um den Modloader einzustellen. Forge, NeoForge, Fabric, und Quilt sind unterstüzt. Wenn Sie eine bestimmte Version des Modloaders verwenden wollen, verwenden Sie die `--loader-version` Option, ansonsten wird die neuste Version verwendet.

Verwenden Sie die `--version, -v` Option um die Spielversion einzustellen, ansonsten wird die neuste Version verwendet. `release` oder `snapshot` sind auch gültige Werte.

Beim Spielstart wird der Launcher versuchen, eine Java-Runtime von Mojang herunterzuladen. Falls es nichts finden kann, müssen Sie die Runtime in der Instanzkonfiguration selbst einrichten.

```sh
cmd-launcher inst create -v 1.21.8 -l fabric CoolInstance
```

**Instanze löschen**  
Wenn Sie eine Instanz löschen möchten, führen Sie den `inst delete` Befehl aus gefolgt von dem Name der Instanz.

### Das Spiel starten

> **WICHTIG!**
> Dieser Launcher ist nicht für Versionen < 1.14 geprüft. Es könnte manchmal für diese Versionen nicht funktioniern, aber ich versuche diese Probleme zu korrigieren!

Um Minecraft zu starten, führen Sie einfach den `start` Befehl gefolgt von dem Name der Instanz aus, die Sie starten wollen.

```bash
cmd-launcher start CoolInstance
```

Um Spieloptionen einzurichten, können Sie Optionen zu den `start` Befehl hinzufügen.

**Gesprächigkeit**  
Um die Gesprächigkeit des Launchers zu ändern, verwenden Sie die `--verbosity` Option. Die mögliche Werte sind:

- `info` - Standard, keine extra Protokollen
- `extra` - Mehr Information beim Spielstart
- `debug` - Debug Information, nützlich für Entwicklung

### Authentifizierung

Wenn Sie im Onlinemodus spielen wollen, müssen Sie ein Microsoft-Konto hinzufügen.

Führen Sie den `auth login` Befehl aus. Der standard Webbrowser wird geöffnet um die Authentifizierung zu starten. Sie können das mit der `--no-browser` Option vermeiden.  
Der Launcher wird automatisch versuchen, im Onlinemodus zu starten, wenn ein Konto angemeldet ist.

Um im Offlinemodus zu spielen, verwenden Sie einfach die `-u, --username <username>` Option beim Spielstart um Ihren Benutzername einzustellen und im Offlinemodus zu starten.

Sie können mit dem `auth logout` Befehl abmelden.

### Instanzkonfiguration

Um Konfigurationswerte zu verändern, öffnen Sie die `instance.toml` Datei in dem Instanzverzeichnis.

Umstellbare Werte sind:

- Spielversion
- Modloader und Modloader Version (wenn nicht Vanilla)
- Spielfenstergröße
- JVM Pfad (wenn leer, eine JVM von Mojang wird heruntergeladen)
- JAR Pfad zu verwenden, statt einen normalen JAR herunterzuladen.
- Extra Java-Argumente
- Minimal- und Maximale Speicherauslastung

Diese Werte können auch in der Command Line überschrieben werden.

**Beispiel `instance.toml` Datei**

```toml
game_version = '1.21.8'
mod_loader = 'fabric'
mod_loader_version = '0.16.14'

[config]
# Path to a Java executable. If blank, a Mojang-provided JVM will be downloaded.
java = '/usr/bin/java'
# Extra arguments to pass to the JVM
java_args = ''
# Path to a custom JAR to use instead of the normal Minecraft client
custom_jar = ''
# Minimum game memory, in MB
min_memory = 512
# Maximum game memory, in MB
max_memory = 4096

# Game window resolution
[config.resolution]
width = 1708
height = 960

```

### Suchen

Der `search` Befehl kann für Minecraft oder Modloader Versionen suchen. Normalerweise sucht er nach Spielversionen, aber er kann auch nach Fabric, Quilt, oder Forge Versionen suchen.

```bash
cmd-launcher search [<query>] [--kind {versions, fabric, quilt, forge}]
```
