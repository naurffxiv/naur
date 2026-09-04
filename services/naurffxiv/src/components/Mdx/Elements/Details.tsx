import { type ReactNode, type ReactElement } from "react";

interface DetailsProps {
  children: ReactNode;
  title: ReactNode;
}

function Details({ children, title }: DetailsProps): ReactElement {
  return (
    <details className="my-4">
      <summary>{title}</summary>

      <div className="p-2">{children}</div>
    </details>
  );
}

export default Details;
