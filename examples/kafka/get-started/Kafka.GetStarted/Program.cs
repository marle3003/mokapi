using System;

if (args.Length == 0)
{
    PrintUsage();
    return;
}

switch (args[0].ToLowerInvariant())
{
    case "producer":
        await Producer.Run();
        break;

    case "consumer":
        Consumer.Run();
        break;

    default:
        Console.WriteLine($"Unknown command: {args[0]}");
        PrintUsage();
        break;
}

static void PrintUsage()
{
    Console.WriteLine("Usage:");
    Console.WriteLine("  dotnet run producer");
    Console.WriteLine("  dotnet run consumer");
}
