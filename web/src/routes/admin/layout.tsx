// src/routes/admin/layout.tsx
import { JSX } from "solid-js";
import AdminLayout from "~/components/admin/Layout";

export default function Layout(props: { children: JSX.Element }) {
  // SolidStart inyectará aquí cualquier página que esté dentro de /admin/
  return <AdminLayout>{props.children}</AdminLayout>;
}