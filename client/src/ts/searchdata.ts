

export interface Status {
    total: number;
    failed: number;
    successful: number;
}

export interface Query {
    query: string;
}

export interface Request {
    query: Query;
    size: number;
    from: number;
    highlight?: any;
    fields?: any;
    facets?: any;
    explain: boolean;
    sort: string[];
    includeLocations: boolean;
    search_after?: any;
    search_before?: any;
}

export interface Hit {
    index: string;
    id: string;
    score: number;
    sort: string[];
}

export interface Search {
    status: Status;
    request: Request;
    hits: Hit[];
    total_hits: number;
    max_score: number;
    took: number;
    facets?: any;
}

export interface File {
    name: string;
    length: number;
}

export interface Torrent {
    infohashHex: string;
    name: string;
    length: number;
    files: File[];
    indexType: string;
}

export interface QueryResponse {
    search: Search;
    torrents: Torrent[];
}



