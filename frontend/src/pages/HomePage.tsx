export default function HomePage() {
  return (
    <div className="home">
      <h1>Bank of Dad</h1>
      <p>Teach your kids about saving and compound interest.</p>
      <a href="/api/auth/google/login" className="btn-google">
        Sign in with Google
      </a>
    </div>
  );
}
