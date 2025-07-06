import { StringSession } from 'telegram/sessions'
import { TelegramClient } from 'telegram'
import { config } from 'dotenv'
// @ts-expect-error input is not a module, but it is in the typescript definition file, so we can ignore this error
// ^^ idk wtf ai is talking about but ok
import input from 'input'

config()

const apiId = Number(process.env.API_ID) || 123456
const apiHash = process.env.API_HASH || ''
const stringSession = new StringSession('') // fill this later with the value from session.save()

;(async () => {
	console.log('Loading interactive session generator...')
	const client = new TelegramClient(stringSession, apiId, apiHash, {
		connectionRetries: 5,
	})
	await client.start({
		phoneNumber: async () => await input.text('Please enter your number: '),
		password: async () => await input.text('Please enter your password: '),
		phoneCode: async () =>
			await input.text('Please enter the code you received: '),
		onError: err => console.log(err),
	})
	console.log('You should now be connected.')
	console.log(client.session.save()) // Save this string to avoid logging in again
	await client.sendMessage('me', {
		message: `generated a new session at ${Date.now().toLocaleString('ru')}`,
	})
})()
