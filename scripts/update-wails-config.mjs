import fs from 'fs'
import path from 'path'

const buildBinDir = path.join('build', 'bin')

const files = fs.readdirSync(buildBinDir)
const setupFile = files.find(
	name => name.startsWith('Void.Presence.Setup.') && name.endsWith('.exe')
)

const base = setupFile.replace(/\.exe$/i, '')
const parts = base.split('.')
const version = parts.slice(3).join('.')
const configPath = 'wails.json'
const config = JSON.parse(fs.readFileSync(configPath, 'utf8'))

config.outputfilename = `Void.Presence.Updates.${version}`

fs.writeFileSync(configPath, JSON.stringify(config, null, 2) + '\n')
