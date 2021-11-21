import Head from 'next/head'

export default function Home() {
  return (
    <div className="w-full h-full bg-red-500 text-white min-h-screen">
      <div className="flex flex-col items-center justify-center min-h-screen food-grid flex-1">
        <Head>
          <title>Mealswipe</title>
          <link rel="icon" href="/favicon.ico" />
        </Head>

        <header className="flex flex-col items-center justify-center w-full px-5 md:px-20 flex-1 text-center">
          <h1 className="text-6xl font-bold drop-shadow">No More Food Fights</h1>
          <h2 className="text-2xl p-5 drop-shadow">Never argue over where to eat again</h2>

          <a href="https://apps.apple.com/us/app/mealswipe-food-made-easy/id1581850924">
          <img src="/download_on_app_store.svg"></img>
          </a>
        </header>

        <footer className="text-center">
          <p>Copyright Disco Studios LLC 2021</p>
        </footer>
      </div>
    </div>

  )
}
