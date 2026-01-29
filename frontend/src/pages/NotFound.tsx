import { Link } from "react-router-dom";

export default function NotFound() {
  return (
    <div className="not-found">
      <h1>This bank doesn't exist</h1>
      <p>
        The family bank you're looking for wasn't found. It may have been
        removed or the URL might be incorrect.
      </p>
      <Link to="/" className="btn-primary">
        Create your own Bank of Dad
      </Link>
    </div>
  );
}
