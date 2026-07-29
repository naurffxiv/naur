"use client";
import Link, { LinkProps } from "next/link";
import { usePathname } from "next/navigation";
import { ReactNode } from "react";

type Props = LinkProps & React.AnchorHTMLAttributes<HTMLAnchorElement>;

/**
 * Simple wrapper for Link that shows a 'selected' style when the
 * current pathname & href match
 * */
export function PathSelectedLink({
  children,
  className,
  ...props
}: Props): ReactNode {
  const pathname = usePathname();
  const isActive = pathname === props.href;

  return (
    <Link
      {...props}
      className={isActive ? `${className} text-[#007EA7]` : className}
    >
      {children}
    </Link>
  );
}
