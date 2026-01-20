import { useEffect, useState } from 'react'

interface MessageResponse {
  message: string
}

function App() {
  const [message, setMessage] = useState<string>('Loading...')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('http://localhost:8001/message')
      .then((response) => {
        if (!response.ok) {
          throw new Error('Failed to fetch message')
        }
        return response.json() as Promise<MessageResponse>
      })
      .then((data) => {
        setMessage(data.message)
      })
      .catch((err) => {
        setError(err.message)
      })
  }, [])

  return (
    <div className="app">
      <h1>Bank of Dad</h1>
      {error ? (
        <p className="error">Error: {error}</p>
      ) : (
        <p className="message">{message}</p>
      )}
    </div>
  )
}

export default App
