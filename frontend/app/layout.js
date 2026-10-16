import "./globals.css";
import { Site } from "../components/site";
export const metadata = {
  title: {
    default: "MICHEAL ENTERPRISE | Your next chapter",
    template: "%s | MICHEAL ENTERPRISE",
  },
  description:
    "Explore our carefully selected vehicles. Speak directly with the MICHEAL ENTERPRISE team for personal dealership service.",
};
export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html:
              "try{document.documentElement.dataset.theme=localStorage.getItem('theme')==='dark'?'dark':'light'}catch(e){}",
          }}
        />
      </head>
      <body>
        <Site>{children}</Site>
      </body>
    </html>
  );
}
