import { Html, Head, Main, NextScript } from 'next/document'
import { Container } from 'react-bootstrap'
 
export default function Document() {
  return (
    <Html data-bs-theme='dark'>
      <Head>
        <link rel="icon" href="data:," />
      </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  )
}