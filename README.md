# cli-cfz

Empaquetador de pases a producción. Toma los archivos que cambió una rama respecto a `main`, los copia
con su estructura de carpetas junto con el `bin` completo y genera una carpeta y un `.zip` con el
nombre de la rama.

## Requisitos

- Go 1.27 o superior (`go version`)
- git en el PATH (`git --version`)
- Windows (probado en Windows 11). No se usa nada exclusivo de Windows, pero no está probado en otros
  sistemas.

## Instalación

```powershell
cd <carpeta del proyecto cli-cfz>
go install ./cmd/cfz
```

Deja `cfz.exe` en `%USERPROFILE%\go\bin`, que debe estar en el PATH. Para verificar:

```powershell
(Get-Command cfz).Source
```

Para actualizar después de cambiar el código, repite `go install ./cmd/cfz`.

## Uso rápido

```powershell
cd <raíz del repositorio>
cfz list
cfz pack
```

Funciona desde cualquier subcarpeta del repositorio: `cfz` busca la raíz y su `cfz.yaml`.

## Comandos

| Comando | Qué hace |
|---|---|
| `cfz` | Modo consola: bienvenida y prompt `cfz>` |
| `cfz init` | Crea un `cfz.yaml` inicial en la raíz del repositorio. `--force` lo reemplaza si ya existe |
| `cfz list` | Muestra qué archivos entrarían al paquete y cuáles se excluyen. No copia nada |
| `cfz pack` | Arma la carpeta y el `.zip` del pase |
| `cfz help` | Lista los comandos |

### `cfz list`

| Opción | Descripción |
|---|---|
| `--branch <rama>` | Compara esa rama contra `ramaBase` en vez de leer los cambios sin commitear |

```
cambios sin commitear en feature/mi-rama

Incluidos (11):
  M  Content/Scripts/Script_Caja.js
  M  Views/Caja/Resumen.cshtml
  ...

Excluidos (18):
  M  appsettings.local.json
  M  Controllers/VentasController.cs
  ...
```

Estados: `A` agregado, `M` modificado, `D` eliminado, `R` renombrado (muestra la ruta anterior).

### `cfz pack`

| Opción | Descripción |
|---|---|
| `--branch <rama>` | Empaqueta esa rama contra `ramaBase`; debe coincidir con la rama actual |
| `--force` | Reemplaza un paquete existente con el mismo nombre |
| `--build` | Ejecuta antes el comando `compilar` del `cfz.yaml`. Si falla, no se crea el paquete |

Sin `--branch` empaqueta lo que cambió respecto al último commit (preparado, sin preparar y archivos
nuevos), estés en la rama que estés. Con `--branch` valida que estés en esa rama y que no haya cambios
sin commitear antes de copiar.

```
Archivos copiados: 11
Archivos del bin:  88
Excluidos:         18
Eliminados:        0

Carpeta: D:\pases\mi-rama_V1
Zip:     D:\pases\mi-rama_V1.zip
```

Resultado:

```
D:\pases\
  mi-rama_V1\        carpeta conservada para revisar
    Content\Scripts\...                  cada archivo con su ruta completa
    Views\...
    bin\...                              bin completo
    ELIMINADOS.txt                       solo si hubo borrados o renombres
  mi-rama_V1.zip
```

### Modo consola

Al escribir `cfz` solo, se limpia la pantalla (como `cls`) y aparece:

```
   ██████╗███████╗███████╗
  ██╔════╝██╔════╝╚══███╔╝
  ██║     █████╗    ███╔╝
  ██║     ██╔══╝   ███╔╝
  ╚██████╗██║     ███████╗
   ╚═════╝╚═╝     ╚══════╝

 cli-cfz v0.1.0 · empaquetador de pases a producción

 Repositorio: D:/proyectos/mi-app
 Rama:        feature/mi-rama
 Config:      cfz.yaml

 Escribe 'help' para ver los comandos o 'exit' para terminar.

▶ cfz>
```

En el prompt se escriben los comandos sin el prefijo (`list`, `pack --force`, `help`).
`cfz list` también funciona. `exit` termina.

### Colores

| Elemento | Estilo |
|---|---|
| Logo y `▶` del prompt | verde |
| `cfz>` del prompt y títulos de `list` | negrita |
| `list`: agregados / modificados / eliminados / renombrados | verde / amarillo / rojo / azul |
| `list`: excluidos | gris |
| `error:` / `ADVERTENCIA:` | rojo / amarillo |

El color se desactiva solo cuando la salida no es una terminal (por ejemplo, `cfz list > lista.txt`)
o cuando existe la variable de entorno `NO_COLOR`. La fuente y el tamaño los define la terminal; si
`▶` se ve como un cuadro, usa una fuente como Cascadia Mono (la predeterminada de Windows Terminal).

## Configuración (`cfz.yaml`)

Cada repositorio necesita un `cfz.yaml` en su raíz. `cfz init` genera uno con `ramaBase: main`,
`salida: ../pases` y los campos opcionales comentados; ajústalo antes de empaquetar.

| Campo | Obligatorio | Uso |
|---|---|---|
| `ramaBase` | sí | Rama contra la que se compara (`main` u `origin/main`) |
| `salida` | sí | Carpeta donde se crean el paquete y el zip. Si es relativa, se toma desde la raíz del repo |
| `bin` | no | Carpeta que se copia completa |
| `compilar` | no | Comando que compila el proyecto (`msbuild App.sln /p:Configuration=Release`). Lo usa `cfz pack --build` desde la raíz del repo |
| `dllPrincipal` | no | Archivo compilado que se compara contra `fuentes` |
| `fuentes` | no | Patrones de archivos que exigen recompilar |
| `sufijo` | no | Texto agregado al nombre del paquete |
| `excluir` | no | Patrones de rutas que nunca entran al paquete |

Los patrones aceptan `**` (cualquier cantidad de carpetas) y no distinguen mayúsculas:
`**/*.cs` excluye todos los `.cs`, `libs/**` excluye una carpeta completa y `Web.config` solo el
archivo de la raíz.

Ejemplo:

```yaml
ramaBase: main
bin: bin
dllPrincipal: bin/App.dll
fuentes:
  - "**/*.cs"
  - "**/*.csproj"
salida: D:/pases
sufijo: _V1
excluir:
  - "**/*.cs"
  - "**/*.csproj"
  - "libs/**"
  - "Web.config"
  - "Web.*.config"
  - "appsettings.local.json"
  - "docs/**"
  - "cfz.yaml"
```

## Errores frecuentes

| Mensaje | Causa | Solución |
|---|---|---|
| `no se encontró ...\cfz.yaml` | El repositorio no tiene configuración | Crear `cfz.yaml` en la raíz |
| `... no es un repositorio git` | Se ejecutó fuera de un repo | Entrar a la carpeta del repositorio |
| `field xxx not found in type config.Config` | Campo mal escrito en `cfz.yaml` | Corregir el nombre del campo |
| `estás en X y pediste empaquetar Y` | `--branch` distinta a la rama actual | `git checkout Y` o quitar `--branch` |
| `hay cambios sin commitear` | Archivos versionados modificados | Commitear o descartar |
| `ya existe ... (usa --force ...)` | Ya se generó ese paquete | `cfz pack --force` |
| `git diff ...: unknown revision` | La rama indicada no existe | Revisar el nombre con `git branch` |
| `ADVERTENCIA: ... vuelve a compilar` | Un `.cs` cambió después de compilar | Recompilar el proyecto y repetir (o usar `cfz pack --build`) |
