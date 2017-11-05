FROM node:alpine
EXPOSE 80
ENV PORT=80
WORKDIR /src
COPY package.json .
RUN npm install
ENTRYPOINT ["npm", "start"]
CMD [] 