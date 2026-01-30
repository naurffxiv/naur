# syntax=docker/dockerfile:1.4

FROM mcr.microsoft.com/dotnet/sdk:10.0 AS build
WORKDIR /src

COPY ["Directory.Packages.props", "."]
COPY ["Directory.Build.props", "."]
COPY ["services/authingway/Authingway.csproj", "services/authingway/"]
COPY ["services/authingway/Directory.Build.props", "services/authingway/"]

RUN --mount=type=cache,target=/root/.nuget/packages \
    dotnet restore "services/authingway/Authingway.csproj"

COPY ["services/authingway/", "services/authingway/"]
WORKDIR "/src/services/authingway"

RUN --mount=type=cache,target=/root/.nuget/packages \
    dotnet publish "Authingway.csproj" \
    -c Release \
    -o /app/publish \
    /p:UseAppHost=false \
    --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:10.0 AS final
WORKDIR /app

RUN apt-get update && \
    apt-get install -y curl && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080
ENV ASPNETCORE_URLS=http://+:8080

COPY --from=build /app/publish .

ENTRYPOINT ["dotnet", "Authingway.dll"]
