import { Html, Head, Main, NextScript } from 'next/document'
import { Container } from 'react-bootstrap'

export default function Document() {
    return (
        <Html data-bs-theme="dark">
            <Head>
                <link rel="icon" href="/favicon.ico" type="image/x-icon" />
                <link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
                <link rel="apple-touch-icon" sizes="180x180" href="/favicon.ico" />
                <meta name="msapplication-TileColor" content="#2b5797" />
            </Head>
            <body>
                <Main />
                <NextScript />
            </body>
        </Html>
    )
}
