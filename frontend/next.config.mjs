export default {
 output: 'standalone',
 poweredByHeader: false,
 async rewrites() { return [{source:'/api/:path*',destination:`${process.env.API_INTERNAL_URL || 'http://localhost:8081'}/api/:path*`},{source:'/uploads/:path*',destination:`${process.env.API_INTERNAL_URL || 'http://localhost:8081'}/uploads/:path*`}]; },
 async headers() {return [{source:'/:path*',headers:[{key:'X-Frame-Options',value:'DENY'},{key:'X-Content-Type-Options',value:'nosniff'},{key:'Referrer-Policy',value:'same-origin'},{key:'Permissions-Policy',value:'camera=(), microphone=(), geolocation=()'},{key:'Content-Security-Policy',value:"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:; font-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"}]}];}
};
