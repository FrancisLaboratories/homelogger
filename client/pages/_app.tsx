import 'bootstrap/dist/css/bootstrap.min.css';
import type { AppProps } from 'next/app'
import Head from 'next/head'
import { useRouter } from 'next/router'
import { APP_VERSION } from '../version'
import { SettingsProvider } from '@/components/SettingsContext'


export default function App({ Component, pageProps }: AppProps) {
  const router = useRouter()

  const getPageName = (p: string) => {
    if (!p || p === '/') return 'Home'
    const first = p.split('/').filter(Boolean)[0] || 'Home'
    return first
      .split(/[-_]/)
      .map(w => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ')
  }

  const pageName = getPageName(router.pathname)

  return (
    <>
      <Head>
        <title>{`HomeLogger | ${pageName}`}</title>
      </Head>
      <SettingsProvider>
        <Component {...pageProps} />
        <footer style={{padding: '12px 0', marginTop: '24px'}}>
          <div style={{textAlign: 'center', color: '#6c757d', fontSize: '0.9rem'}}>
            <span style={{display: 'inline-flex', alignItems: 'center', gap: 8}}>
              <img src="/logoname.png" alt="HomeLogger" style={{height: 22}} />
              <span>v{APP_VERSION}</span>
              <a href="https://github.com/FrancisLaboratories/homelogger" target="_blank" rel="noopener noreferrer" style={{display: 'inline-flex', alignItems: 'center', color: '#6c757d', textDecoration: 'none', marginLeft: 8}}>
                <img src="/github.png" alt="GitHub" style={{width: 16, height: 16, marginLeft: 6}} />
              </a>
            </span>
          </div>
          <div style={{textAlign: 'center', color: '#6c757d', fontSize: '0.8rem', marginTop: '4px'}}>
            Made with &#x2665; in Detroit
          </div>
        </footer>
      </SettingsProvider>
    </>
  )
}
