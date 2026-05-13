---
name: sdd-full
description: >
  Run the COMPLETE Spec-Driven Development pipeline end-to-end with ZERO user
  confirmations. Recorre las 14 fases (web) o las 11 (non-web) del DAG SDD
  encadenadas, sin pausar para preguntar, eligiendo defaults razonables en
  cualquier punto que normalmente pediría input. Es `/sdd full` con `--auto`
  forzado como única forma de operación.
  Trigger: /sdd-full [change], /sdd-full --auto [change], "lanza sdd full",
  "sdd full auto", "corre el pipeline completo sin preguntar".
license: MIT
metadata:
  author: gentleman-programming
  version: "1.0"
  parent_orchestrator: ~/.claude/skills/sdd/SKILL.md
---

# SDD Full — Auto Pipeline (No-Confirm)

Esto es un wrapper sobre el orquestador `sdd`. La diferencia con `/sdd full`
es **una sola y crítica**: aquí NUNCA se le pide NADA al usuario. Cero
confirmaciones. Cero preguntas. Cero "¿procedo?". Cero "¿estás seguro?".
Cero `AskUserQuestion`. Cero "¿me das permiso para…?".

El usuario lanza `/sdd-full <change>` y vuelve cuando termina (o cuando
falla). Punto.

## Contrato `--auto` — alcance EXPLÍCITO

`--auto` cubre los siguientes tipos de "pausa", todos prohibidos:

1. **`AskUserQuestion`** — prohibido completamente.
2. **Preguntas inline en texto** — prohibido. Nada de "¿quieres que…?",
   "¿procedo con X?", "antes de continuar, confirmame…", "¿te parece bien
   si…?". Si te encontrás escribiendo esa frase, **eliminala y seguí**.
3. **Permisos de herramienta** — el usuario corre con `permissionMode`
   `acceptEdits` o `bypassPermissions`. Asumí que Bash, browse, edits,
   writes, fetches y MCP están todos pre-aprobados. Si una herramienta
   exige permiso runtime y te corta, registralo en `auto-decisions` como
   `tool_blocked` y seguí con la siguiente fase (no preguntes al usuario).
4. **Cambios de directorio** — `cd` libre dentro del workspace del
   proyecto (cwd actual y descendientes). NO preguntás antes.
5. **Lanzar el dev server o browse para pruebas** — adelante, sin pedir.
   Si el puerto está ocupado, matá el proceso anterior y reintentá.
6. **Crear archivos nuevos, ejecutar tests, instalar deps locales** —
   adelante, sin pedir.
7. **Decisiones de diseño/scope dentro del DAG** — adoptás la opción
   más conservadora por defecto y la registrás en `auto-decisions`.

## Contrato heredado a sub-agentes

Cuando el orquestador `sdd` lance cada fase vía Task, **debe pasar a cada
sub-agente** este bloque literal en el prompt:

```
AUTO-MODE: este lanzamiento es parte de /sdd-full --auto. Reglas DURAS:
- NO formules AskUserQuestion. Está prohibido.
- NO escribas preguntas inline ("¿procedo?", "¿confirmás?", "¿estás
  seguro?"). Está prohibido.
- NO pidas permiso para cd, browse, bash, edit, write, fetch ni ninguna
  herramienta. El usuario corre en acceptEdits/bypassPermissions.
- Si una herramienta te corta por permiso runtime: registralo en
  auto-decisions como tool_blocked y devolvé status=PARTIAL con lo que
  hayas podido hacer. NO frenes el DAG por eso.
- Si una decisión requiere preferencia, elegí la opción más conservadora
  y documentala en el artefacto de salida bajo "auto-decisions".
- Solo se permite parar el DAG en: FAIL real (test rojo, build roto,
  fase imposible), o acción destructiva no-local (push --force, rm -rf
  fuera de workspace, cambios en CI/CD). En esos dos casos, devolvé
  status=BLOCKED con motivo concreto.
- RTK obligatorio: todo Bash debe pasar por rtk (hook PreToolUse del
  usuario ya lo reescribe transparente; si lanzás un comando que no
  pasa por hook, llamalo como `rtk <cmd>` explícito). Política del
  usuario, no negociable.
```

---

## Defaults para fases interactivas

| Fase                | Pausa normal en `/sdd full`              | Default en `sdd-full --auto`                                                                |
|---------------------|------------------------------------------|---------------------------------------------------------------------------------------------|
| explore (office-hours) | Preguntas del modo Startup/Builder    | Saltar el modo conversacional. Genera explore-doc directo desde el prompt del usuario + lectura del repo. |
| design (plan-eng-review) | Recorrer dimensiones interactivamente | Recorre TODAS las dimensiones, asigna scores, no pregunta — adopta la recomendación más segura por defecto. |
| review (gstack-review)   | "ASK items" para taste decisions       | Trata cada ASK como NIT y sigue. Documenta los ASK no resueltos en el review-report como `pending_user_review`. |
| qa (gstack-qa)           | Fix-loop puede pedir confirmar fixes   | Tier `Quick`. Aplica fixes obvios sin preguntar. Marca fixes ambiguos como `deferred`.       |
| ship                     | Confirma versión/CHANGELOG             | Acepta autopick de versión. CHANGELOG generado automáticamente. NO push a main sin que CI pase. |

**Nota de seguridad**: incluso en `--auto`, NO se aprueban acciones
destructivas no-locales. Si una fase requiere `git push --force`, `rm -rf`
fuera del workspace, o cambios en CI/CD, se **detiene la pipeline** y se
reporta al usuario. "Auto" significa "sin preguntar trivialidades", no
"saltar guardrails".

---

## DAG ejecutado

**Web project** (14 pasos):
```
init → explore → propose → spec → design → tasks → apply → browse-check →
verify → review → qa → ship → archive → canary
```

**Non-web project** (11 pasos):
```
init → explore → propose → spec → design → tasks → apply → verify → review →
ship → archive
```

La detección web/non-web ocurre en `init` (ver `sdd-init/SKILL.md`).

---

## Flow de ejecución

1. **Cargar contexto del orquestador**: leer `~/.claude/skills/sdd/SKILL.md`
   secciones "Dependency DAG", "Node Table", "Sub-Agent Launch Protocol",
   "Model Routing".
2. **Resolver `<change>`**: si el usuario pasa nombre, usarlo. Si no, generar
   uno desde el último prompt (`mem_save_prompt` ya lo tiene) en kebab-case.
3. **Init phase**: si no hay `sdd-init/{project}` en engram, lanzar
   `sdd-init` primero. Detectar web/non-web.
4. **Bucle del DAG**: por cada nodo en orden topológico, lanzar Task con el
   contrato de no-pausa inyectado. Esperar `status` en el envelope SDD.
   - `OK`/`PASS`/`PASS_WITH_WARNINGS` → siguiente nodo.
   - `FAIL`/`BLOCKED` → **parar** la pipeline, reportar al usuario el motivo
     exacto y el artefacto de salida. No reintentar automáticamente.
5. **Progreso visible**: después de cada fase, una línea estilo:
   ```
   [4/14] design ✅ — 8 dimensiones, score 7.2
   [5/14] tasks ← RUNNING…
   ```
6. **Engram bridging**: cada artefacto bajo `topic_key: sdd/{change}/{phase}`,
   igual que `/sdd full` (ver `_shared/engram-convention.md`).
7. **Resumen final**: cuando llegue a `archive` (o `canary` en web), imprimir
   tabla de las fases ejecutadas con sus artefactos y `auto-decisions`
   acumuladas. El usuario revisa **post-hoc**, no durante.

---

## Invocación

```
/sdd-full [change]
/sdd-full --auto [change]    # --auto es redundante (ya es el default)
```

Aliases aceptados (texto natural):
- "lanza sdd full"
- "corre el pipeline completo"
- "sdd full auto <change>"
- "ejecuta todas las fases sin preguntar"

---

## Diferencia explícita con `/sdd full` (subcomando del skill `sdd`)

| Comportamiento                          | `/sdd full` | `sdd-full` (esta skill)        |
|-----------------------------------------|-------------|--------------------------------|
| Encadena las 14 fases                   | Sí          | Sí                             |
| Pausa en office-hours, plan-eng-review, review ASK | Sí (línea 82 del orquestador) | NO — usa defaults    |
| Acción destructiva no-local sin permiso | No          | No (mismo guardrail)           |
| Reporta `auto-decisions` al final       | No          | Sí (artefacto agregado)        |
| Aborta en FAIL/BLOCKED                  | Sí          | Sí                             |

Si querés el flujo con pausas, usá `/sdd full`. Si querés todo seguido,
usá `/sdd-full` (esta).
