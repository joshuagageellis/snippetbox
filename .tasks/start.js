import 'dotenv/config'
import bs from 'browser-sync';

export default function start() {
	const server = bs.create();

	console.log(`Server running on port ${process.env.PORT}`);

	server.init({
		proxy: `localhost:${process.env.PORT}`,
		open: false,
		delay: 1000,
		files: [
			'**/*',
		]
	});
}

start();