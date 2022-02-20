<?php

namespace DesignPatterns\Structural\Flyweight;

/**
 * The Flyweight Factory stores both the Context and Flyweight objects,
 * effectively hiding any notion of the Flyweight pattern from the client.
 */
class CatDataBase
{
    /**
     * The list of cat objects (Contexts).
     */
    private $cats = [];

    /**
     * The list of cat variations (Flyweights).
     */
    private $variations = [];

    /**
     * When adding a cat to the database, we look for an existing cat variation
     * first.
     */
    public function addCat(
        string $name,
        string $age,
        string $owner,
        string $breed,
        string $image,
        string $color,
        string $texture,
        string $fur,
        string $size
    ): array
    {
        $variation = $this->getVariation($breed, $image, $color, $texture, $fur, $size);
        $this->cats[] = new Cat($name, $age, $owner, $variation);

        return $this->cats;
    }

    /**
     * Return an existing variation (Flyweight) by given data or create a new
     * one if it doesn't exist yet.
     *
     * @param string $breed
     * @param string $image
     * @param string $color
     * @param string $texture
     * @param string $fur
     * @param string $size
     * @return CatVariation
     */
    public function getVariation(
        string $breed,
        string $image,
        string $color,
        string $texture,
        string $fur,
        string $size
    ): CatVariation
    {
        $key = $this->getKey(get_defined_vars());

        if (!isset($this->variations[$key])) {
            $this->variations[$key] = new CatVariation($breed, $image, $color, $texture, $fur, $size);
        }

        return $this->variations[$key];
    }

    /**
     * This function helps to generate unique array keys.
     *
     * @param array $data
     * @return string
     */
    private function getKey(array $data): string
    {
        return md5(implode("_", $data));
    }

    /**
     * Look for a cat in the database using the given query parameters.
     */
    public function findCat(array $query)
    {
        foreach ($this->cats as $cat) {
            if ($cat->matches($query)) {
                return $cat;
            }
        }
        throw new \Exception("CatDataBase: Sorry, your query does not yield any results.");
    }
}